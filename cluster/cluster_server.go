package cluster

import (
	"encoding/binary"
	"mqtt_server/MQTT_Server_Go/config"
	"mqtt_server/MQTT_Server_Go/log"
	"net"
	"strconv"
	"strings"
	"time"
)

const (
	TIME_1_SECOND = 100
	TIME_1_MINUTE = 100 * 60
)

const (
	CLUSTER_HEART_MSG = 0x0001
)

type ClusterRoutingMsg struct {
	MsgType uint16
	MsgBody interface{}
}

var (
	clusterServerNodes []*ClusterNodeInfo
)

func buildClusterHeartMsg() []byte {
	msg := make([]byte, 4)
	binary.BigEndian.PutUint16(msg[0:2], CLUSTER_HEART_MSG)
	binary.BigEndian.PutUint16(msg[2:4], 0)
	return msg
}

func processClusterNodeMsg(data []byte, length int, addr *net.UDPAddr, ch chan ClusterRoutingMsg) {
	msgType := binary.BigEndian.Uint16(data[0:2])
	msgLen := binary.BigEndian.Uint16(data[2:4])
	if length < int(msgLen+4) {
		log.LogPrint(log.LOG_WARNING, "recv data error %v", msgLen)
		return
	}
	switch msgType {
	case CLUSTER_HEART_MSG:
		var msg ClusterRoutingMsg
		var remoteAddr net.UDPAddr
		msg.MsgType = CLUSTER_HEART_MSG
		remoteAddr = *addr
		msg.MsgBody = remoteAddr
		ch <- msg
	}
}

func startClusterServerRecv(conn *net.UDPConn, ch chan ClusterRoutingMsg) {

	defer conn.Close()

	for {
		data := make([]byte, 4096)
		n, raddr, err := conn.ReadFromUDP(data)
		if err != nil {
			log.LogPrint(log.LOG_PANIC, "Failed To Read UDP Msg, Error: %v", err)
			break
		}
		log.LogPrint(log.LOG_DEBUG, "Read UDP Msg, Len: %v, from: %v", n, raddr)
		if n < 4 {
			log.LogPrint(log.LOG_DEBUG, "data is invalid")
		} else {
			processClusterNodeMsg(data, n, raddr, ch)
		}
	}
}

func startClusterServerCycly(conn *net.UDPConn, ch chan ClusterRoutingMsg) {
	timeEscape100MS := 0
	timeEscape1S := 0
	ticker := time.NewTicker(time.Millisecond * 100)

	for {
		var msg ClusterRoutingMsg
		select {
		case <-ticker.C:
			timeEscape100MS++
			if timeEscape100MS == TIME_1_SECOND {
				timeEscape1S++
				for _, serverNode := range clusterServerNodes {
					_, alive := serverNode.KeepAliveTimeEscape1Sec()
					if !alive {
						log.LogPrint(log.LOG_WARNING, "cluster server node [%s:%d] maybe exception remove it", serverNode.Ip, serverNode.Port)
					} else {
						if timeEscape1S == 50 {
							msg := buildClusterHeartMsg()
							serverNode.SendMsgToNode(conn, msg)
							timeEscape1S = 0
						}
					}
				}
				timeEscape100MS = 0
			}
		case msg = <-ch:
			switch msg.MsgType {
			case CLUSTER_HEART_MSG:
				raddr, ok := msg.MsgBody.(net.UDPAddr)
				if ok {
					for _, serverNode := range clusterServerNodes {
						if serverNode.Ip == raddr.IP.To4().String() && serverNode.Port == uint16(raddr.Port) {
							serverNode.UpdateKeepAliveTimeout()
							break
						}
					}
				}
			}
		}
	}
	ticker.Stop()
}

func createClusterSocket(address string) (*net.UDPConn, error) {
	udpAddr, err := net.ResolveUDPAddr("udp", address)
	if err != nil {
		return nil, err
	}
	conn, err := net.ListenUDP("udp", udpAddr)
	if err != nil {
		return nil, err
	}
	return conn, nil
}

func StartClusterServerNode() int {
	ClusterPort, _ := config.ReadConfigValueInt("cluster", "port")
	address := "0.0.0.0:" + strconv.Itoa(ClusterPort)
	conn, err := createClusterSocket(address)
	if err != nil {
		return 0
	}
	cluster_nodes, err := config.ReadConfigValueString("cluster", "coservers")
	if err != nil {
		return 0
	}
	serverNodes := strings.Split(cluster_nodes, "/")
	for _, nodeAddr := range serverNodes {
		log.LogPrint(log.LOG_DEBUG, "Get one server node [%s]", nodeAddr)
		node := CreateClusterNode(nodeAddr)
		clusterServerNodes = append(clusterServerNodes, node)
	}
	cluster_ch := make(chan ClusterRoutingMsg)
	go startClusterServerRecv(conn, cluster_ch)
	go startClusterServerCycly(conn, cluster_ch)
	return 1
}

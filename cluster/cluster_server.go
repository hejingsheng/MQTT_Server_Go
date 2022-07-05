package cluster

import (
	"encoding/binary"
	"encoding/json"
	"mqtt_server/MQTT_Server_Go/config"
	"mqtt_server/MQTT_Server_Go/log"
	"mqtt_server/MQTT_Server_Go/process"
	"mqtt_server/MQTT_Server_Go/protocol_stack"
	"net"
	"strconv"
	"strings"
	"time"
)

const (
	TIME_1_SECOND = 10
	TIME_1_MINUTE = 10 * 60
)

const (
	CLUSTER_HEART_MSG = 0x0001
	CLUSTER_PUBLIC_MSG = 0x0002
	CLUSTER_LOGIN_MSG = 0x0003
)

type ClusterRoutingMsg struct {
	MsgType uint16
	MsgBody interface{}
}

var (
	clusterServerNodes []*ClusterNodeInfo
	Cluster_Ch chan ClusterRoutingMsg
)

func buildClusterHeartMsg() []byte {
	msg := make([]byte, 4)
	binary.BigEndian.PutUint16(msg[0:2], CLUSTER_HEART_MSG)
	binary.BigEndian.PutUint16(msg[2:4], 0)
	return msg
}

func buildClusterPublishMsg(pubdata []byte) []byte {
	msg := make([]byte, 4)
	length := len(pubdata)
	binary.BigEndian.PutUint16(msg[0:2], CLUSTER_PUBLIC_MSG)
	binary.BigEndian.PutUint16(msg[2:4], uint16(length))
	msg = append(msg, pubdata...)
	return msg
}

func processClusterNodeMsg(data []byte, length int, addr *net.UDPAddr, ch chan ClusterRoutingMsg, ch1 chan process.DispatchRoutinMsg) {
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
	case CLUSTER_PUBLIC_MSG:
		msgBody := data[4:]
		var pubData protocol_stack.MQTTPacketPublishData
		err := json.Unmarshal(msgBody, &pubData)
		if err != nil {
			log.LogPrint(log.LOG_ERROR, "public msg data error")
		} else {
			var dispatch_msg process.DispatchRoutinMsg
			dispatch_msg.MsgType = process.MSG_CLUSTER_PUBLISH
			dispatch_msg.MsgBody = pubData
			dispatch_msg.MsgFrom = addr.String()
			ch1 <- dispatch_msg
		}
	}
}

func startClusterServerRecv(conn *net.UDPConn, ch chan ClusterRoutingMsg, dispatchCh chan process.DispatchRoutinMsg) {

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
			processClusterNodeMsg(data, n, raddr, ch, dispatchCh)
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
							log.LogPrint(log.LOG_INFO, "update [%s:%d] cluster server node", serverNode.Ip, serverNode.Port)
							serverNode.UpdateKeepAliveTimeout()
							break
						}
					}
				}
			case CLUSTER_PUBLIC_MSG:
				publishData := msg.MsgBody.(protocol_stack.MQTTPacketPublishData)
				jsonPub, err := json.Marshal(publishData)
				if err != nil {
					log.LogPrint(log.LOG_WARNING, "json error")
				} else {
					for _, serverNode := range clusterServerNodes {
						_, alive := serverNode.KeepAliveTimeEscape1Sec()
						if alive {
							msg := buildClusterPublishMsg(jsonPub)
							serverNode.SendMsgToNode(conn, msg)
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

func StartClusterServerNode(dispatchCh chan process.DispatchRoutinMsg) int {
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
	Cluster_Ch = make(chan ClusterRoutingMsg)
	go startClusterServerRecv(conn, Cluster_Ch, dispatchCh)
	go startClusterServerCycly(conn, Cluster_Ch)
	return 1
}

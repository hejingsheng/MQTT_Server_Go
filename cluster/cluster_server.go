package cluster

import (
	"encoding/binary"
	"encoding/json"
	"mqtt_server/MQTT_Server_Go/config"
	"mqtt_server/MQTT_Server_Go/log"
	"mqtt_server/MQTT_Server_Go/protocol_stack"
	"net"
	"strconv"
	"strings"
	"time"
)

const (
	MAX_UDP_DATA_LEN = 1024
)

const (
	TIME_1_SECOND = 10
	TIME_1_MINUTE = 10 * 60
)

const (
	CLUSTER_HEART_MSG = 0x0001
	CLUSTER_PUBLIC_MSG = 0x0002
	CLUSTER_LOGIN_MSG = 0x0003
	CLUSTER_CLIENT_INFO_MSG = 0x0004
)

const (
	CLUSTER_MSG_PUBLISH_DISPATCH = 100
	CLUSTER_MSG_LOGIN_NOTIFY = 101
	CLUSTER_MSG_LOGIN_INFO_NOTIFY = 102
)

type ClusterClientInfo struct {
	ClientId string                                   `json:"clientid"`
	Subinfo map[string]uint8                          `json:"subinfo"`
	OfflineMsg []protocol_stack.MQTTPacketPublishData `json:"offlinemsg"`
}

type RoutingCommunicateMsg struct {
	MsgType uint16
	MsgFrom string
	MsgBody interface{}
}

var (
	clusterServerNodes []*ClusterNodeInfo
	Cluster_Ch chan RoutingCommunicateMsg
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

func buildClusterLoginMsg(clientId string) []byte {
	msg := make([]byte, 4)
	length := len(clientId)
	binary.BigEndian.PutUint16(msg[0:2], CLUSTER_LOGIN_MSG)
	binary.BigEndian.PutUint16(msg[2:4], uint16(length))
	msg = append(msg, []byte(clientId)...)
	return msg
}

func buildClusterLoginInfoMsg(info []byte) [][]byte {
	var msg_silce [][]byte
	length := len(info)
	index := 0
	lastPkt := false
	var msg []byte
	for {
		if index < length {
			leng := 0
			if length - index > MAX_UDP_DATA_LEN {
				leng = MAX_UDP_DATA_LEN
			} else {
				leng = length - index
				lastPkt = true
			}
			msg = make([]byte, 5)
			binary.BigEndian.PutUint16(msg[0:2], CLUSTER_CLIENT_INFO_MSG)
			binary.BigEndian.PutUint16(msg[2:4], uint16(leng+1))
			if lastPkt {
				msg[4] = 1
			} else {
				msg[4] = 0
			}
			msg = append(msg, info[index:index+leng]...)
			index += leng;
		} else {
			break;
		}
	}
	return msg_silce
}

var clientInfoMsg []byte
func processClusterNodeMsg(data []byte, length int, addr *net.UDPAddr, ch chan RoutingCommunicateMsg, ch1 chan RoutingCommunicateMsg) {
	msgType := binary.BigEndian.Uint16(data[0:2])
	msgLen := binary.BigEndian.Uint16(data[2:4])
	if length != int(msgLen+4) {
		log.LogPrint(log.LOG_WARNING, "recv data error %v", msgLen)
		return
	}
	switch msgType {
	case CLUSTER_HEART_MSG:
		var msg RoutingCommunicateMsg
		var remoteAddr net.UDPAddr
		msg.MsgType = CLUSTER_HEART_MSG
		remoteAddr = *addr
		msg.MsgBody = remoteAddr
		ch <- msg
	case CLUSTER_PUBLIC_MSG:
		msgBody := data[4:length]
		var pubData protocol_stack.MQTTPacketPublishData
		err := json.Unmarshal(msgBody, &pubData)
		if err != nil {
			log.LogPrint(log.LOG_ERROR, "public msg data error")
		} else {
			var dispatch_msg RoutingCommunicateMsg
			dispatch_msg.MsgType = CLUSTER_MSG_PUBLISH_DISPATCH
			dispatch_msg.MsgBody = pubData
			dispatch_msg.MsgFrom = addr.String()
			ch1 <- dispatch_msg
		}
	case CLUSTER_LOGIN_MSG:
		loginClientId := string(data[4:length])
		var login_msg RoutingCommunicateMsg
		login_msg.MsgType = CLUSTER_MSG_LOGIN_NOTIFY
		login_msg.MsgBody = loginClientId
		login_msg.MsgFrom = addr.String()
		ch1 <- login_msg
	case CLUSTER_CLIENT_INFO_MSG:
		lastPkt := data[4]
		if lastPkt == 1 {
			clientInfoMsg = append(clientInfoMsg, data[5:length]...)
			var info ClusterClientInfo
			err := json.Unmarshal(clientInfoMsg, &info)
			if err != nil {
				log.LogPrint(log.LOG_ERROR, "clientinfo msg data error")
			} else {
				var login_info_msg RoutingCommunicateMsg
				login_info_msg.MsgType = CLUSTER_MSG_LOGIN_INFO_NOTIFY
				login_info_msg.MsgBody = info
				login_info_msg.MsgFrom = addr.String()
				ch1 <- login_info_msg
			}
			clientInfoMsg = []byte{}
		} else {
			clientInfoMsg = append(clientInfoMsg, data[5:length]...)
		}
	}
}

func startClusterServerRecv(conn *net.UDPConn, ch chan RoutingCommunicateMsg, ch1 chan RoutingCommunicateMsg) {

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
			processClusterNodeMsg(data, n, raddr, ch, ch1)
		}
	}
}

func startClusterServerCycly(conn *net.UDPConn, ch chan RoutingCommunicateMsg) {
	timeEscape100MS := 0
	timeEscape1S := 0
	ticker := time.NewTicker(time.Millisecond * 100)

	for {
		var msg RoutingCommunicateMsg
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
			case CLUSTER_LOGIN_MSG:
				clientId := msg.MsgBody.(string)
				for _, serverNode := range clusterServerNodes {
					_, alive := serverNode.KeepAliveTimeEscape1Sec()
					if alive {
						msg := buildClusterLoginMsg(clientId)
						serverNode.SendMsgToNode(conn, msg)
					}
				}
			case CLUSTER_CLIENT_INFO_MSG:
				clientInfo := msg.MsgBody.(ClusterClientInfo)
				to := msg.MsgFrom
				log.LogPrint(log.LOG_INFO, "need sync client [%s] info to cluster server node %s", clientInfo.ClientId, to)
				index := strings.Index(to, ":")
				ip := to[0:index]
				port, _ := strconv.Atoi(ip[index+1 : len(ip)])
				jsonStr, err := json.Marshal(clientInfo)
				if err != nil {
					log.LogPrint(log.LOG_WARNING, "json error")
				} else {
					for _, serverNode := range clusterServerNodes {
						if serverNode.Ip == ip && serverNode.Port == uint16(port) {
							_, alive := serverNode.KeepAliveTimeEscape1Sec()
							if alive {
								log.LogPrint(log.LOG_INFO, "sync info to [%s:%d] cluster server node", serverNode.Ip, serverNode.Port)
								msg_list := buildClusterLoginInfoMsg(jsonStr)
								for _, msg := range msg_list {
									serverNode.SendMsgToNode(conn, msg)
								}
							} else {
								log.LogPrint(log.LOG_ERROR, "sync info to [%s:%d] cluster node but it is offline")
							}
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

func StartClusterServerNode(dispatch_ch chan RoutingCommunicateMsg) int {
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
	Cluster_Ch = make(chan RoutingCommunicateMsg)
	go startClusterServerRecv(conn, Cluster_Ch, dispatch_ch)
	go startClusterServerCycly(conn, Cluster_Ch)
	return 1
}

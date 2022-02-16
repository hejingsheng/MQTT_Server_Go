package process

import (
	"github.com/mqtt_server/MQTT_Server_Go/log"
	"github.com/mqtt_server/MQTT_Server_Go/protocol_stack"
	"github.com/mqtt_server/MQTT_Server_Go/utils"
	"net"
	"sync"
)

type MqttClientInfo struct {
	clientId     string
	username     string
	password     string
	pingperiod   uint16
	clearSession uint8
	loginSuccess byte
	conn  net.Conn

	//subInfo []protocol_stack.MqttSubInfo
	subInfo   map[string]uint8 // subscribe topic  key = topic  value = qos
	pubStatus map[uint16]uint8 // publish msg status key = packetId value = status for qos1 and qos2 status is puback pubrec pubrel pubcomp
	offlineMsg []protocol_stack.MQTTPubInfo // save offline msg
}

var (
	gloablClientsMap     map[string]*MqttClientInfo
	globalClientsMapLock sync.Mutex
	index                int = 1
)

func init() {
	gloablClientsMap = make(map[string]*MqttClientInfo, 1)
}

func ClientProcess(localconn net.Conn) {

	var routinId string  // use client struct address as thread id
	var mqttClientData *MqttClientInfo = nil
	//mqttClientData := MqttClientInfo{
	//	clientId:     "",
	//	username:     "",
	//	password:     "",
	//	pingperiod:   0,
	//	clearSession: 0,
	//	loginSuccess: 0,
	//}

	routinId = utils.GeneralRoutinId(12)
	log.LogPrint(log.LOG_INFO,  routinId, "start a tcp client")
	defer func() {
		log.LogPrint(log.LOG_INFO, routinId, "one client close exit process")
		if mqttClientData.clearSession == 1 { // client request clear session in connect data
			globalClientsMapLock.Lock()
			delete(gloablClientsMap, mqttClientData.clientId)
			globalClientsMapLock.Unlock()
		} else {
			mqttClientData.loginSuccess = 0
		}
		localconn.Close()
		utils.RemoveRoutinId(routinId)
	}()

	for {
		var read_buf []byte = make([]byte, 1024)
		num, err := localconn.Read(read_buf)
		if err != nil {
			log.LogPrint(log.LOG_ERROR, routinId, "read error")
			break
		} else {
			log.LogPrint(log.LOG_DEBUG, routinId, "client read %d data %v", num, read_buf)
			mqttMsgType, _ := protocol_stack.MqttPacket_ParseFixHeader(read_buf)

			if mqttMsgType != -1 {
				switch mqttMsgType {
				case protocol_stack.CONNECT:
					var conndata protocol_stack.MQTTPacketConnectData
					conndata.MQTTDeserialize_connect(read_buf, num)
					var write_buf []byte = make([]byte, 0)
					num, ret := conndata.MQTTSeserialize_connack(&write_buf)
					log.LogPrint(log.LOG_DEBUG, routinId, "write %d data, ret %d", num, ret)
					if ret == 0 {
						var username, password, clientId string
						pingperiod, clearSession := conndata.MQTT_GetConnectInfo(&username, &password, &clientId)
						_, ok := gloablClientsMap[clientId]
						if ok {
							// this clientId have login and not clear session
							mqttClientData = gloablClientsMap[clientId]
						} else {
							// first login or clear session need new client struct
							mqttClientData = new(MqttClientInfo)
							mqttClientData.subInfo = make(map[string]uint8)
							mqttClientData.pubStatus = make(map[uint16]uint8)
							mqttClientData.offlineMsg = make([]protocol_stack.MQTTPubInfo, 0)
						}
						mqttClientData.loginSuccess = 1
						mqttClientData.clientId = clientId
						mqttClientData.username = username
						mqttClientData.password = password
						mqttClientData.pingperiod = pingperiod
						mqttClientData.clearSession = clearSession
						mqttClientData.conn = localconn
						globalClientsMapLock.Lock()
						gloablClientsMap[mqttClientData.clientId] = mqttClientData
						globalClientsMapLock.Unlock()
						log.LogPrint(log.LOG_INFO, routinId, "this client [%s:%s] login success", mqttClientData.clientId, mqttClientData.username)
						localconn.Write(write_buf)
					} else {
						localconn.Write(write_buf)
						localconn.Close()
					}
				case protocol_stack.DISCONNECT:
					log.LogPrint(log.LOG_INFO, routinId, "client [%s] send disconnect req", mqttClientData.clientId)
					localconn.Close()
				case protocol_stack.SUBSCRIBE:
					if mqttClientData.loginSuccess != 1 {
						log.LogPrint(log.LOG_ERROR, routinId, "client not login")
					} else {
						var subscribeData protocol_stack.MQTTPacketSubscribeData
						topicNum := subscribeData.MQTTDeserialize_subscribe(read_buf, num)
						log.LogPrint(log.LOG_INFO, routinId, "client [%s] subscribe %d topic", mqttClientData.clientId, topicNum)
						var write_buf []byte = make([]byte, 0)
						subscribeData.MQTTGetSubInfo(mqttClientData.subInfo, topicNum)
						subscribeData.MQTTSeserialize_suback(&write_buf, topicNum)
						localconn.Write(write_buf)
					}
				case protocol_stack.UNSUBSCRIBE:
					if mqttClientData.loginSuccess != 1 {
						log.LogPrint(log.LOG_ERROR, routinId, "client not login")
					} else {
						var unSubscribeData protocol_stack.MQTTPacketUnSubscribeData
						topicNum := unSubscribeData.MQTTDeserialize_unsubscribe(read_buf, num)
						log.LogPrint(log.LOG_INFO, routinId, "client [%s] unsubscribe %d topic", mqttClientData.clientId, topicNum)
						var write_buf []byte = make([]byte, 0)
						unSubscribeData.MQTTRemoveUnSubTopic(mqttClientData.subInfo)
						unSubscribeData.MQTTSeserialize_unsuback(&write_buf)
						localconn.Write(write_buf)
					}
				case protocol_stack.PUBLISH:
					if mqttClientData.loginSuccess != 1 {
						log.LogPrint(log.LOG_ERROR, routinId, "client not login")
					} else {
						var publishData protocol_stack.MQTTPacketPublishData
						var pubInfo protocol_stack.MQTTPubInfo
						ret := publishData.MQTTDeserialize_publish(read_buf, num)
						publishData.MQTTGetPublishInfo(&pubInfo)
						log.LogPrint(log.LOG_INFO, routinId, "client [%s] public msg %d, public %s,payload %d", mqttClientData.clientId, ret, pubInfo.Topic, pubInfo.Qos)
						var write_buf []byte = read_buf[:num]
						for _, client := range gloablClientsMap {
							//if conn == localconn{
							//	continue
							//}
							for topic := range client.subInfo {
								if topic == pubInfo.Topic {
									client.conn.Write(write_buf)
									if pubInfo.Qos == 1 {
										mqttClientData.pubStatus[pubInfo.PacketId] = protocol_stack.PUBACK
									} else if pubInfo.Qos == 2 {
										mqttClientData.pubStatus[pubInfo.PacketId] = protocol_stack.PUBREC
									} else if pubInfo.Qos == 0 {

									} else {
										log.LogPrint(log.LOG_ERROR, routinId, "error Qos level %d", pubInfo.Qos)
									}
								}
							}
						}
					}
				case protocol_stack.PUBACK:
					var publishData protocol_stack.MQTTPacketPublishData
					var pubInfo protocol_stack.MQTTPubInfo
					publishData.MQTTDeserialize_puback(read_buf, num)
					publishData.MQTTGetPublishInfo(&pubInfo)
					var write_buf []byte = read_buf[:num]
					for _, client := range gloablClientsMap {
						//if conn == localconn{
						//	continue
						//}
						status, ok := client.pubStatus[pubInfo.PacketId]
						if ok && status == protocol_stack.PUBACK {
							client.conn.Write(write_buf)
							delete(client.pubStatus, pubInfo.PacketId)
						} else {
							log.LogPrint(log.LOG_WARNING, routinId, "not find client or client status error, drop it")
						}
					}
				case protocol_stack.PUBREC:
					var publishData protocol_stack.MQTTPacketPublishData
					var pubInfo protocol_stack.MQTTPubInfo
					publishData.MQTTDeserialize_pubrec(read_buf, num)
					publishData.MQTTGetPublishInfo(&pubInfo)
					var write_buf []byte = read_buf[:num]
					for _, client := range gloablClientsMap {
						//if conn == localconn{
						//	continue
						//}
						status, ok := client.pubStatus[pubInfo.PacketId]
						if ok && status == protocol_stack.PUBREC {
							mqttClientData.pubStatus[pubInfo.PacketId] = protocol_stack.PUBREL
							client.conn.Write(write_buf)
						} else {
							log.LogPrint(log.LOG_WARNING, routinId, "not find client or client status error, drop it")
						}
					}
				case protocol_stack.PUBREL:
					var publishData protocol_stack.MQTTPacketPublishData
					var pubInfo protocol_stack.MQTTPubInfo
					publishData.MQTTDeserialize_pubrel(read_buf, num)
					publishData.MQTTGetPublishInfo(&pubInfo)
					var write_buf []byte = read_buf[:num]
					for _, client := range gloablClientsMap {
						//if conn == localconn{
						//	continue
						//}
						status, ok := client.pubStatus[pubInfo.PacketId]
						if ok && status == protocol_stack.PUBREL {
							mqttClientData.pubStatus[pubInfo.PacketId] = protocol_stack.PUBCOMP
							client.conn.Write(write_buf)
						} else {
							log.LogPrint(log.LOG_WARNING, routinId, "not find client or client status error, drop it")
						}
					}
				case protocol_stack.PUBCOMP:
					var publishData protocol_stack.MQTTPacketPublishData
					var pubInfo protocol_stack.MQTTPubInfo
					publishData.MQTTDeserialize_pubcomp(read_buf, num)
					publishData.MQTTGetPublishInfo(&pubInfo)
					var write_buf []byte = read_buf[:num]
					for _, client := range gloablClientsMap {
						//if conn == localconn{
						//	continue
						//}
						status, ok := client.pubStatus[pubInfo.PacketId]
						if ok && status == protocol_stack.PUBCOMP {
							log.LogPrint(log.LOG_DEBUG, routinId,"client [%s] current status %d, remote status %d", mqttClientData.clientId, mqttClientData.pubStatus[pubInfo.PacketId], client.pubStatus[pubInfo.PacketId])
							mqttClientData.pubStatus[pubInfo.PacketId] = protocol_stack.NONE
							client.pubStatus[pubInfo.PacketId] = protocol_stack.NONE
							delete(mqttClientData.pubStatus, pubInfo.PacketId)
							delete(client.pubStatus, pubInfo.PacketId)
							client.conn.Write(write_buf)
						} else {
							log.LogPrint(log.LOG_WARNING, routinId, "not find client or client status error, drop it")
						}
					}
				case protocol_stack.PINGREQ:
					var pingpongdata protocol_stack.MQTTPacketPingPongData
					pingpongdata.MQTTDeserialize_ping(read_buf, num)
					var write_buf []byte = make([]byte, 0)
					num = pingpongdata.MQTTSeserialize_pong(&write_buf)
					localconn.Write(write_buf)
				}
			}

		}
	}
}

package process

import (
	"mqtt_server/MQTT_Server_Go/log"
	"mqtt_server/MQTT_Server_Go/protocol_stack"
	"mqtt_server/MQTT_Server_Go/utils"
	"net"
	"sync"
)

type MqttClientInfo struct {
	ClientId     string
	Username     string
	password     string
	pingperiod   uint16
	clearSession uint8
	LoginSuccess byte
	conn         net.Conn

	heartTimeout uint16

	//subInfo []protocol_stack.MqttSubInfo
	SubInfo    map[string]uint8                       // subscribe topic  key = topic  value = qos
	PubStatus  map[uint16]uint8                       // publish msg status key = packetId value = status for qos1 and qos2 status is puback pubrec pubrel pubcomp
	OfflineMsg []protocol_stack.MQTTPacketPublishData // save offline msg
	CacheMsg   map[uint16]protocol_stack.MQTTPacketPublishData // save not process complete msg for qos2 key=packetid value=publishmsg
}

func (mqttClientData *MqttClientInfo) GeneralUniquePacketId() uint16 {
	var packetId uint16
	for {
		packetId = utils.GeneralPacketID()
		_, ok := mqttClientData.PubStatus[packetId]
		if !ok {
			return packetId
		}
	}
}

func (mqttClientData *MqttClientInfo) ProcessSubscribe(routinId string, read_buf []byte, num int) int {
	if mqttClientData.LoginSuccess != 1 {
		log.LogPrint(log.LOG_ERROR, routinId, "client not login")
		return -1
	} else {
		var subscribeData protocol_stack.MQTTPacketSubscribeData
		topicNum := subscribeData.MQTTDeserialize_subscribe(read_buf, num)
		log.LogPrint(log.LOG_INFO, routinId, "client [%s] subscribe %d topic", mqttClientData.ClientId, topicNum)
		subscribeData.MQTTGetSubInfo(mqttClientData.SubInfo, topicNum)
		var write_buf []byte = make([]byte, 0)
		subscribeData.MQTTSeserialize_suback(&write_buf, topicNum)
		mqttClientData.conn.Write(write_buf)
	}
	return 0
}

func (mqttClientData *MqttClientInfo) ProcessUnSubscribe(routinId string, read_buf []byte, num int) int {
	if mqttClientData.LoginSuccess != 1 {
		log.LogPrint(log.LOG_ERROR, routinId, "client not login")
	} else {
		var unSubscribeData protocol_stack.MQTTPacketUnSubscribeData
		topicNum := unSubscribeData.MQTTDeserialize_unsubscribe(read_buf, num)
		log.LogPrint(log.LOG_INFO, routinId, "client [%s] unsubscribe %d topic", mqttClientData.ClientId, topicNum)
		var write_buf []byte = make([]byte, 0)
		unSubscribeData.MQTTRemoveUnSubTopic(mqttClientData.SubInfo)
		unSubscribeData.MQTTSeserialize_unsuback(&write_buf)
		mqttClientData.conn.Write(write_buf)
	}
	return 0
}

func (mqttClientData *MqttClientInfo) ProcessPublish(routinId string, read_buf []byte, num int) int {
	if mqttClientData.LoginSuccess != 1 {
		log.LogPrint(log.LOG_ERROR, routinId, "client not login")
	} else {
		var publishData protocol_stack.MQTTPacketPublishData
		publishData.MQTTDeserialize_publish(read_buf, num)
		//publishData.MQTTGetPublishInfo(&pubInfo)
		log.LogPrint(log.LOG_INFO, routinId, "client [%s] pid %d, public %s,payload %d", mqttClientData.ClientId, publishData.PacketId, publishData.Topic, publishData.Qos)
		var write_buf []byte = read_buf[:num]
		var write_buf_ack []byte = make([]byte, 0) //
		//var write_buf_offline []byte = make([]byte, 0) //
		if publishData.Qos == 1 {
			publishData.MQTTSeserialize_puback(&write_buf_ack)
			mqttClientData.conn.Write(write_buf_ack)
		} else if publishData.Qos == 2 {
			// server recv qos2 msg send pubrec msg wait client send pubrel
			log.LogPrint(log.LOG_DEBUG, routinId, "client [%s] send pubrec pid %d", mqttClientData.ClientId, publishData.PacketId)
			publishData.MQTTSeserialize_pubrec(&write_buf_ack)
			mqttClientData.PubStatus[publishData.PacketId] = protocol_stack.PUBREL
			mqttClientData.CacheMsg[publishData.PacketId] = publishData
			mqttClientData.conn.Write(write_buf_ack)
		} else {
			// Qos = 0
		}
		// dispatch msg need Regenerate packet id for qos1 and qos2 because different client maybe generate same packet id
		// so we need generate packet id by server to provice different packet id for each client
		for _, client := range GloablClientsMap {
			//if conn == localconn{
			//	continue
			//}
			for topic := range client.SubInfo {
				if topic == publishData.Topic {
					publishData.PacketId = client.GeneralUniquePacketId()
					if client.LoginSuccess == 1 {
						if publishData.Qos == 1 {
							write_buf = []byte{}
							publishData.MQTTSeserialize_publish(&write_buf)
							client.conn.Write(write_buf)
							client.PubStatus[publishData.PacketId] = protocol_stack.PUBACK
							log.LogPrint(log.LOG_INFO, routinId, "client[%s]->client[%s],pid=%d", mqttClientData.ClientId, client.ClientId, publishData.PacketId)
						} else if publishData.Qos == 2 {
							// qos2 msg not complete do not do anything
							//publishData.PacketId = client.GeneralUniquePacketId()
							//client.CacheMsg[publishData.PacketId] = publishData
							//mqttClientData.PubStatus[publishData.PacketId] = protocol_stack.PUBREC
						} else if publishData.Qos == 0 {
							client.conn.Write(write_buf)
						} else {
							log.LogPrint(log.LOG_ERROR, routinId, "error Qos level %d", publishData.Qos)
						}
					} else {
						if publishData.Qos > 0 {
							log.LogPrint(log.LOG_INFO, routinId, "client [%s] is offline, but qos(%d) > 0", client.ClientId, publishData.Qos)
							if publishData.Qos == 1 {
								client.OfflineMsg = append(client.OfflineMsg, publishData)
							} else if publishData.Qos == 2 {
								log.LogPrint(log.LOG_INFO, routinId, "client [%s] is offline msg qos(2) need wait complete", client.ClientId)
							} else {
								log.LogPrint(log.LOG_ERROR, routinId, "error Qos level %d", publishData.Qos)
							}
						}
					}
				}
			}
		}
	}
	return 0
}

func (mqttClientData *MqttClientInfo) ProcessPublishAck(routinId string, read_buf []byte, num int) int {
	var publishData protocol_stack.MQTTPacketPublishData
	publishData.MQTTDeserialize_puback(read_buf, num)
	status, ok := mqttClientData.PubStatus[publishData.PacketId]
	if ok && status == protocol_stack.PUBACK {
		log.LogPrint(log.LOG_WARNING, routinId, "client[%s] recv puback for pid %d", mqttClientData.ClientId, publishData.PacketId)
		delete(mqttClientData.PubStatus, publishData.PacketId)
	} else {
		log.LogPrint(log.LOG_WARNING, routinId, "not find client or client status error, drop it")
	}
	return 0
}

func (mqttClientData *MqttClientInfo) ProcessPublishRec(routinId string, read_buf []byte, num int) int {
	var publishData protocol_stack.MQTTPacketPublishData
	publishData.MQTTDeserialize_pubrec(read_buf, num)
	status, ok := mqttClientData.PubStatus[publishData.PacketId]
	if ok && status == protocol_stack.PUBREC {
		log.LogPrint(log.LOG_DEBUG, routinId, "client[%s] send pubrel for pid %d qos2 msg wait pubcomp", mqttClientData.ClientId, publishData.PacketId)
		var write_buf_ack []byte = make([]byte, 0)
		publishData.MQTTSeserialize_pubrel(&write_buf_ack)
		mqttClientData.PubStatus[publishData.PacketId] = protocol_stack.PUBCOMP
		mqttClientData.conn.Write(write_buf_ack)
	} else {
		log.LogPrint(log.LOG_WARNING, routinId, "not find client or client status error, drop it")
	}
	return 0
}

func (mqttClientData *MqttClientInfo) ProcessPublishRel(routinId string, read_buf []byte, num int) int {
	var publishData protocol_stack.MQTTPacketPublishData
	publishData.MQTTDeserialize_pubrel(read_buf, num)
	//var write_buf []byte = read_buf[:num]
	var write_buf []byte = make([]byte, 0)
	status, ok := mqttClientData.PubStatus[publishData.PacketId]
	if ok && status == protocol_stack.PUBREL {
		publishData.MQTTSeserialize_pubcomp(&write_buf)
		delete(mqttClientData.PubStatus, publishData.PacketId)
		mqttClientData.conn.Write(write_buf)
		//mqttClientData.PubStatus[publishData.PacketId] = protocol_stack.NONE
		log.LogPrint(log.LOG_DEBUG, routinId, "client[%s] send pubcomp for pid %d qos2 msg complete", mqttClientData.ClientId, publishData.PacketId)
		msg, ok := mqttClientData.CacheMsg[publishData.PacketId]
		if ok {
			delete(mqttClientData.CacheMsg, publishData.PacketId)
			log.LogPrint(log.LOG_DEBUG, routinId, "client[%s] pid %d qos2 msg complete delete it from cachemsg start dispatch", mqttClientData.ClientId, publishData.PacketId)
			// Qos2 msg complete for publish client need dispatch this Qos2 msg to other client
			for _, client := range GloablClientsMap {
				for topic := range client.SubInfo {
					if topic == msg.Topic {
						msg.PacketId = client.GeneralUniquePacketId()
						log.LogPrint(log.LOG_DEBUG, routinId, "client[%s]->client[%s] send pid %d qos2 msg", mqttClientData.ClientId, client.ClientId, msg.PacketId)
						if client.LoginSuccess == 1 {
							write_buf = []byte{}
							msg.MQTTSeserialize_publish(&write_buf)
							client.CacheMsg[msg.PacketId] = msg
							client.PubStatus[msg.PacketId] = protocol_stack.PUBREC
							client.conn.Write(write_buf)
						} else {
							log.LogPrint(log.LOG_INFO, routinId, "client [%s] msg qos2 complete but client [%s] offline need save offline", mqttClientData.ClientId, client.ClientId)
							client.OfflineMsg = append(client.OfflineMsg, msg)
						}
					}
				}
			}
		} else {
			log.LogPrint(log.LOG_WARNING, routinId, "not find client msg")
		}
	} else {
		log.LogPrint(log.LOG_WARNING, routinId, "not find client or client status error, drop it")
		return 0;
	}
	return 0
}
func (mqttClientData *MqttClientInfo) ProcessPublishComp(routinId string, read_buf []byte, num int) int {
	var publishData protocol_stack.MQTTPacketPublishData
	publishData.MQTTDeserialize_pubcomp(read_buf, num)
	status, ok := mqttClientData.PubStatus[publishData.PacketId]
	if ok && status == protocol_stack.PUBCOMP {
		log.LogPrint(log.LOG_DEBUG, routinId, "client[%s] recv pubcomp for pid %d qos2 msg complete delete it", mqttClientData.ClientId, publishData.PacketId)
		mqttClientData.PubStatus[publishData.PacketId] = protocol_stack.NONE
		delete(mqttClientData.CacheMsg, publishData.PacketId)
		delete(mqttClientData.PubStatus, publishData.PacketId)
	} else {
		log.LogPrint(log.LOG_WARNING, routinId, "not find client or client status error, drop it")
		return 0;
	}
	return 0
}

var (
	GloablClientsMap     map[string]*MqttClientInfo
	globalClientsMapLock sync.Mutex
	index                int = 1
)

func init() {
	GloablClientsMap = make(map[string]*MqttClientInfo, 1)
}

func ClientProcess(localconn net.Conn, cycleCh chan CycleRoutinMsg) {

	var routinId string // use client struct address as thread id
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
	log.LogPrint(log.LOG_INFO, routinId, "start a mqtt client")
	defer func() {
		log.LogPrint(log.LOG_INFO, routinId, "client [%s:%d] close exit process", mqttClientData.ClientId, mqttClientData.clearSession)
		if mqttClientData.clearSession == 1 { // client request clear session in connect data
			//globalClientsMapLock.Lock()
			//delete(gloablClientsMap, mqttClientData.clientId)
			//globalClientsMapLock.Unlock()
			var msg CycleRoutinMsg
			msg.MsgType = MSG_DEL_CLIENT
			msg.MsgBody = mqttClientData.ClientId
			cycleCh <- msg
		} else {
			mqttClientData.LoginSuccess = 0
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
			//log.LogPrint(log.LOG_DEBUG, routinId, "client read %d data %v", num, read_buf)
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
						//var username, password, clientId string
						//pingperiod, clearSession := conndata.MQTT_GetConnectInfo(&username, &password, &clientId)
						_, ok := GloablClientsMap[conndata.ClientID]
						if ok {
							// this clientId have login and not clear session
							mqttClientData = GloablClientsMap[conndata.ClientID]
						} else {
							// first login or clear session need new client struct
							mqttClientData = new(MqttClientInfo)
							mqttClientData.SubInfo = make(map[string]uint8)
							mqttClientData.PubStatus = make(map[uint16]uint8)
							mqttClientData.OfflineMsg = make([]protocol_stack.MQTTPacketPublishData, 0)
							mqttClientData.CacheMsg = make(map[uint16]protocol_stack.MQTTPacketPublishData)
						}
						mqttClientData.LoginSuccess = 1
						mqttClientData.heartTimeout = conndata.KeepAliveInterval
						mqttClientData.ClientId = conndata.ClientID
						mqttClientData.Username = conndata.Username
						mqttClientData.password = conndata.Password
						mqttClientData.pingperiod = conndata.KeepAliveInterval
						mqttClientData.clearSession = conndata.Cleansession
						mqttClientData.conn = localconn
						log.LogPrint(log.LOG_INFO, routinId, "this client [%s:%s] login success", mqttClientData.ClientId, mqttClientData.Username)
						localconn.Write(write_buf)
						var msg CycleRoutinMsg
						msg.MsgType = MSG_ADD_CLIENT
						msg.MsgBody = mqttClientData
						cycleCh <- msg
						for _, offlineMsg := range mqttClientData.OfflineMsg {
							//var msg *protocol_stack.MQTTPacketPublishData = &offlineMsg
							write_buf = []byte{}
							offlineMsg.MQTTSeserialize_publish(&write_buf)
							if offlineMsg.Qos == 2 {
								mqttClientData.PubStatus[offlineMsg.PacketId] = protocol_stack.PUBREC
							}
							localconn.Write(write_buf)
							//msg.MQTTSeserialize_publish()
						}
						mqttClientData.OfflineMsg = []protocol_stack.MQTTPacketPublishData{}
					} else {
						localconn.Write(write_buf)
						localconn.Close()
					}
				case protocol_stack.DISCONNECT:
					log.LogPrint(log.LOG_INFO, routinId, "client [%s] send disconnect req", mqttClientData.ClientId)
					localconn.Close()
				case protocol_stack.SUBSCRIBE:
					mqttClientData.ProcessSubscribe(routinId, read_buf, num)
				case protocol_stack.UNSUBSCRIBE:
					mqttClientData.ProcessUnSubscribe(routinId, read_buf, num)
				case protocol_stack.PUBLISH:
					mqttClientData.ProcessPublish(routinId, read_buf, num)
				case protocol_stack.PUBACK:
					mqttClientData.ProcessPublishAck(routinId, read_buf, num)
				case protocol_stack.PUBREC:
					mqttClientData.ProcessPublishRec(routinId, read_buf, num)
				case protocol_stack.PUBREL:
					mqttClientData.ProcessPublishRel(routinId, read_buf, num)
				case protocol_stack.PUBCOMP:
					mqttClientData.ProcessPublishComp(routinId, read_buf, num)
				case protocol_stack.PINGREQ:
					mqttClientData.heartTimeout = mqttClientData.pingperiod
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

package process

import (
	"mqtt_server/MQTT_Server_Go/log"
	"mqtt_server/MQTT_Server_Go/protocol_stack"
	"mqtt_server/MQTT_Server_Go/utils"
	"net"
	"strconv"
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
	SubInfo    sync.Map //[string]uint8                       // subscribe topic  key = topic  value = qos
	PubStatus  sync.Map //[uint16]uint8                       // publish msg status key = packetId value = status for qos1 and qos2 status is puback pubrec pubrel pubcomp
	OfflineMsg []protocol_stack.MQTTPacketPublishData // save offline msg
	CacheMsg   sync.Map//[uint16]protocol_stack.MQTTPacketPublishData // save not process complete msg for qos2 key=packetid value=publishmsg
}

func (mqttClientData *MqttClientInfo) GeneralUniquePacketId() uint16 {
	var packetId uint16
	for {
		packetId = utils.GeneralPacketID()
		//_, ok := mqttClientData.PubStatus[packetId]
		_, ok := mqttClientData.PubStatus.Load(packetId)
		if !ok {
			return packetId
		}
	}
}

func (mqttClientData *MqttClientInfo) ProcessSubscribe(routinId string, read_buf []byte, num int) int {
	if mqttClientData.LoginSuccess != 1 {
		log.LogPrint(log.LOG_ERROR, "[%s] client not login", routinId)
		return -1
	} else {
		var subscribeData protocol_stack.MQTTPacketSubscribeData
		topicNum := subscribeData.MQTTDeserialize_subscribe(read_buf, num)
		log.LogPrint(log.LOG_INFO, "[%s] client [%s] subscribe %d topic", routinId, mqttClientData.ClientId, topicNum)
		subInfo := make(map[string]uint8)
		subscribeData.MQTTGetSubInfo(subInfo, topicNum)
		for topic, qos := range subInfo {
			mqttClientData.SubInfo.Store(topic, qos)
		}
		var write_buf []byte = make([]byte, 0)
		subscribeData.MQTTSeserialize_suback(&write_buf, topicNum)
		mqttClientData.conn.Write(write_buf)
	}
	return 0
}

func (mqttClientData *MqttClientInfo) ProcessUnSubscribe(routinId string, read_buf []byte, num int) int {
	if mqttClientData.LoginSuccess != 1 {
		log.LogPrint(log.LOG_ERROR, "[%s] client not login", routinId)
	} else {
		var unSubscribeData protocol_stack.MQTTPacketUnSubscribeData
		topicNum := unSubscribeData.MQTTDeserialize_unsubscribe(read_buf, num)
		log.LogPrint(log.LOG_INFO, "[%s] client [%s] unsubscribe %d topic", routinId, mqttClientData.ClientId, topicNum)
		var write_buf []byte = make([]byte, 0)
		unSubscribeData.MQTTRemoveUnSubTopic(&mqttClientData.SubInfo)
		unSubscribeData.MQTTSeserialize_unsuback(&write_buf)
		mqttClientData.conn.Write(write_buf)
	}
	return 0
}

func (mqttClientData *MqttClientInfo) ProcessPublish(routinId string, read_buf []byte, num int, cycleCh chan CycleRoutinMsg) int {
	if mqttClientData.LoginSuccess != 1 {
		log.LogPrint(log.LOG_ERROR, "[%s] client not login", routinId)
	} else {
		var publishData protocol_stack.MQTTPacketPublishData
		publishData.MQTTDeserialize_publish(read_buf, num)
		//publishData.MQTTGetPublishInfo(&pubInfo)
		log.LogPrint(log.LOG_INFO, "[%s] client [%s] pid %d, public %s,payload %d", routinId, mqttClientData.ClientId, publishData.PacketId, publishData.Topic, publishData.Qos)
		//var write_buf []byte = read_buf[:num]
		var write_buf_ack []byte = make([]byte, 0) //
		//var write_buf_offline []byte = make([]byte, 0) //
		if publishData.Qos == 1 {
			publishData.MQTTSeserialize_puback(&write_buf_ack)
			mqttClientData.conn.Write(write_buf_ack)
		} else if publishData.Qos == 2 {
			// server recv qos2 msg send pubrec msg wait client send pubrel
			log.LogPrint(log.LOG_DEBUG, "[%s] client [%s] send pubrec pid %d", routinId, mqttClientData.ClientId, publishData.PacketId)
			publishData.MQTTSeserialize_pubrec(&write_buf_ack)
			//mqttClientData.PubStatus[publishData.PacketId] = protocol_stack.PUBREL
			//mqttClientData.CacheMsg[publishData.PacketId] = publishData
			mqttClientData.PubStatus.Store(publishData.PacketId, protocol_stack.PUBREL)
			mqttClientData.CacheMsg.Store(publishData.PacketId, publishData)
			mqttClientData.conn.Write(write_buf_ack)
		} else {
			// Qos = 0
		}
		// dispatch msg need Regenerate packet id for qos1 and qos2 because different client maybe generate same packet id
		// so we need generate packet id by server to provice different packet id for each client
		var msg_ch CycleRoutinMsg
		msg_ch.MsgType = MSG_PUB_CLIENT
		msg_ch.MsgFrom = mqttClientData.ClientId
		msg_ch.MsgBody = publishData
		cycleCh <- msg_ch
	}
	return 0
}

func (mqttClientData *MqttClientInfo) ProcessPublishAck(routinId string, read_buf []byte, num int) int {
	var publishData protocol_stack.MQTTPacketPublishData
	publishData.MQTTDeserialize_puback(read_buf, num)
	//status, ok := mqttClientData.PubStatus[publishData.PacketId]
	status, ok := mqttClientData.PubStatus.Load(publishData.PacketId)
	if ok && status == protocol_stack.PUBACK {
		log.LogPrint(log.LOG_WARNING, "[%s] client[%s] recv puback for pid %d", routinId, mqttClientData.ClientId, publishData.PacketId)
		//delete(mqttClientData.PubStatus, publishData.PacketId)
		mqttClientData.PubStatus.Delete(publishData.PacketId)
	} else {
		log.LogPrint(log.LOG_WARNING, "[%s] not find client or client status error, drop it", routinId)
	}
	return 0
}

func (mqttClientData *MqttClientInfo) ProcessPublishRec(routinId string, read_buf []byte, num int) int {
	var publishData protocol_stack.MQTTPacketPublishData
	publishData.MQTTDeserialize_pubrec(read_buf, num)
	//status, ok := mqttClientData.PubStatus[publishData.PacketId]
	status, ok := mqttClientData.PubStatus.Load(publishData.PacketId)
	if ok && status == protocol_stack.PUBREC {
		log.LogPrint(log.LOG_DEBUG, "[%s] client[%s] send pubrel for pid %d qos2 msg wait pubcomp", routinId, mqttClientData.ClientId, publishData.PacketId)
		var write_buf_ack []byte = make([]byte, 0)
		publishData.MQTTSeserialize_pubrel(&write_buf_ack)
		//mqttClientData.PubStatus[publishData.PacketId] = protocol_stack.PUBCOMP
		mqttClientData.PubStatus.Store(publishData.PacketId, protocol_stack.PUBCOMP)
		mqttClientData.conn.Write(write_buf_ack)
	} else {
		log.LogPrint(log.LOG_WARNING, "[%s] not find client or client status error, drop it", routinId)
	}
	return 0
}

func (mqttClientData *MqttClientInfo) ProcessPublishRel(routinId string, read_buf []byte, num int, cycleCh chan CycleRoutinMsg) int {
	var publishData protocol_stack.MQTTPacketPublishData
	publishData.MQTTDeserialize_pubrel(read_buf, num)
	//var write_buf []byte = read_buf[:num]
	var write_buf []byte = make([]byte, 0)
	//status, ok := mqttClientData.PubStatus[publishData.PacketId]
	status, ok := mqttClientData.PubStatus.Load(publishData.PacketId)
	if ok && status == protocol_stack.PUBREL {
		publishData.MQTTSeserialize_pubcomp(&write_buf)
		//delete(mqttClientData.PubStatus, publishData.PacketId)
		mqttClientData.PubStatus.Delete(publishData.PacketId)
		mqttClientData.conn.Write(write_buf)
		//mqttClientData.PubStatus[publishData.PacketId] = protocol_stack.NONE
		log.LogPrint(log.LOG_DEBUG, "[%s] client[%s] send pubcomp for pid %d qos2 msg complete", routinId, mqttClientData.ClientId, publishData.PacketId)
		//msg, ok := mqttClientData.CacheMsg[publishData.PacketId]
		msg, ok := mqttClientData.CacheMsg.Load(publishData.PacketId)
		if ok {
			//delete(mqttClientData.CacheMsg, publishData.PacketId)
			mqttClientData.CacheMsg.Delete(publishData.PacketId)
			log.LogPrint(log.LOG_DEBUG, "[%s] client[%s] pid %d qos2 msg complete delete it from cachemsg start dispatch", routinId, mqttClientData.ClientId, publishData.PacketId)
			// Qos2 msg complete for publish client need dispatch this Qos2 msg to other client
			var msg_ch CycleRoutinMsg
			msg_ch.MsgType = MSG_REL_CLIENT
			msg_ch.MsgFrom = mqttClientData.ClientId
			msg_ch.MsgBody = msg
			cycleCh <- msg_ch
		} else {
			log.LogPrint(log.LOG_WARNING, "[%s] not find client msg", routinId)
		}
	} else {
		log.LogPrint(log.LOG_WARNING, "[%s] not find client or client status error, drop it", routinId)
		return 0;
	}
	return 0
}
func (mqttClientData *MqttClientInfo) ProcessPublishComp(routinId string, read_buf []byte, num int) int {
	var publishData protocol_stack.MQTTPacketPublishData
	publishData.MQTTDeserialize_pubcomp(read_buf, num)
	//status, ok := mqttClientData.PubStatus[publishData.PacketId]
	status, ok := mqttClientData.PubStatus.Load(publishData.PacketId)
	if ok && status == protocol_stack.PUBCOMP {
		log.LogPrint(log.LOG_DEBUG, "[%s] client[%s] recv pubcomp for pid %d qos2 msg complete delete it", routinId, mqttClientData.ClientId, publishData.PacketId)
		//mqttClientData.PubStatus[publishData.PacketId] = protocol_stack.NONE
		mqttClientData.PubStatus.Store(publishData.PacketId, protocol_stack.NONE)
		//delete(mqttClientData.CacheMsg, publishData.PacketId)
		//delete(mqttClientData.PubStatus, publishData.PacketId)
		mqttClientData.CacheMsg.Delete(publishData.PacketId)
		mqttClientData.PubStatus.Delete(publishData.PacketId)
	} else {
		log.LogPrint(log.LOG_WARNING, "[%s] not find client or client status error, drop it", routinId)
		return 0;
	}
	return 0
}

var (
	GloablClientsMap     map[string]*MqttClientInfo
	GlobalClientsMapLock sync.RWMutex
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

	remoteaddr := utils.Addr2Int(localconn.RemoteAddr().String())
	routinId = strconv.FormatInt(remoteaddr, 16)
	routinId = ("0x"+routinId)
	log.LogPrint(log.LOG_INFO, "[%s] start a mqtt client", routinId)
	defer func() {
		log.LogPrint(log.LOG_INFO, "[%s] client [%s:%d] close exit process", routinId, mqttClientData.ClientId, mqttClientData.clearSession)
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
		//utils.RemoveRoutinId(routinId)
	}()

	for {
		var read_buf []byte = make([]byte, 1024)
		num, err := localconn.Read(read_buf)
		if err != nil {
			log.LogPrint(log.LOG_ERROR, "[%s] read error", routinId)
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
					log.LogPrint(log.LOG_DEBUG, "[%s] write %d data, ret %d", routinId, num, ret)
					if ret == 0 {
						//var username, password, clientId string
						//pingperiod, clearSession := conndata.MQTT_GetConnectInfo(&username, &password, &clientId)
						GlobalClientsMapLock.RLock()
						_, ok := GloablClientsMap[conndata.ClientID]
						if ok {
							// this clientId have login and not clear session
							mqttClientData = GloablClientsMap[conndata.ClientID]
						} else {
							// first login or clear session need new client struct
							mqttClientData = new(MqttClientInfo)
							//mqttClientData.SubInfo = make(map[string]uint8)
							//mqttClientData.PubStatus = make(map[uint16]uint8)
							mqttClientData.OfflineMsg = make([]protocol_stack.MQTTPacketPublishData, 0)
							//mqttClientData.CacheMsg = make(map[uint16]protocol_stack.MQTTPacketPublishData)
						}
						mqttClientData.LoginSuccess = 1
						mqttClientData.heartTimeout = conndata.KeepAliveInterval
						mqttClientData.ClientId = conndata.ClientID
						mqttClientData.Username = conndata.Username
						mqttClientData.password = conndata.Password
						mqttClientData.pingperiod = conndata.KeepAliveInterval
						mqttClientData.clearSession = conndata.Cleansession
						mqttClientData.conn = localconn
						log.LogPrint(log.LOG_INFO, "[%s] this client [%s:%s] login success", routinId, mqttClientData.ClientId, mqttClientData.Username)
						localconn.Write(write_buf)
						for _, offlineMsg := range mqttClientData.OfflineMsg {
							//var msg *protocol_stack.MQTTPacketPublishData = &offlineMsg
							write_buf = []byte{}
							offlineMsg.MQTTSeserialize_publish(&write_buf)
							if offlineMsg.Qos == 2 {
								//mqttClientData.PubStatus[offlineMsg.PacketId] = protocol_stack.PUBREC
								mqttClientData.PubStatus.Store(offlineMsg.PacketId, protocol_stack.PUBREC)
							}
							localconn.Write(write_buf)
							//msg.MQTTSeserialize_publish()
						}
						mqttClientData.OfflineMsg = []protocol_stack.MQTTPacketPublishData{}
						GlobalClientsMapLock.RUnlock()
						var msg CycleRoutinMsg
						msg.MsgType = MSG_ADD_CLIENT
						msg.MsgBody = mqttClientData
						cycleCh <- msg
					} else {
						localconn.Write(write_buf)
						localconn.Close()
					}
				case protocol_stack.DISCONNECT:
					log.LogPrint(log.LOG_INFO, "[%s] client [%s] send disconnect req", routinId, mqttClientData.ClientId)
					localconn.Close()
				case protocol_stack.SUBSCRIBE:
					mqttClientData.ProcessSubscribe(routinId, read_buf, num)
				case protocol_stack.UNSUBSCRIBE:
					mqttClientData.ProcessUnSubscribe(routinId, read_buf, num)
				case protocol_stack.PUBLISH:
					mqttClientData.ProcessPublish(routinId, read_buf, num, cycleCh)
				case protocol_stack.PUBACK:
					mqttClientData.ProcessPublishAck(routinId, read_buf, num)
				case protocol_stack.PUBREC:
					mqttClientData.ProcessPublishRec(routinId, read_buf, num)
				case protocol_stack.PUBREL:
					mqttClientData.ProcessPublishRel(routinId, read_buf, num, cycleCh)
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

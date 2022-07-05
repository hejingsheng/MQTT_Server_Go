package process

import (
	"mqtt_server/MQTT_Server_Go/log"
	"mqtt_server/MQTT_Server_Go/protocol_stack"
	"time"
)

const (
	MSG_ADD_CLIENT = iota
	MSG_DEL_CLIENT
	MSG_PUB_CLIENT
	MSG_REL_CLIENT
)

type CycleRoutinMsg struct {
	MsgType uint16
	MsgFrom string
	MsgBody interface{}
}

func processMsg(routinId string, msg CycleRoutinMsg) {
	switch msg.MsgType {
	case MSG_ADD_CLIENT:
		mqttClient,ok := msg.MsgBody.(*MqttClientInfo)
		if ok {
			log.LogPrint(log.LOG_DEBUG,"[%s] read from cycle channal one client login", routinId)
			GloablClientsMap[mqttClient.ClientId] = mqttClient
		} else {
			log.LogPrint(log.LOG_ERROR, "[%s] add client msg data is error", routinId)
		}
	case MSG_DEL_CLIENT:
		clientId,ok := msg.MsgBody.(string)
		if ok {
			delete(GloablClientsMap, clientId)
		} else {
			log.LogPrint(log.LOG_INFO, "[%s] del client msg data is error", routinId)
		}
	case MSG_PUB_CLIENT:
		publishData := msg.MsgBody.(protocol_stack.MQTTPacketPublishData)
		from := msg.MsgFrom
		write_buf := []byte{}
		for _, client := range GloablClientsMap {
			//if conn == localconn{
			//	continue
			//}
			for topic := range client.SubInfo {
				if topic == publishData.Topic {
					publishData.PacketId = client.GeneralUniquePacketId()
					if client.LoginSuccess == 1 {
						if publishData.Qos == 1 {
							publishData.MQTTSeserialize_publish(&write_buf)
							client.PubStatus[publishData.PacketId] = protocol_stack.PUBACK
							client.conn.Write(write_buf)
							log.LogPrint(log.LOG_INFO, "[%s] client[%s]->client[%s],pid=%d", routinId, from, client.ClientId, publishData.PacketId)
						} else if publishData.Qos == 2 {
							// qos2 msg not complete do not do anything
							//publishData.PacketId = client.GeneralUniquePacketId()
							//client.CacheMsg[publishData.PacketId] = publishData
							//mqttClientData.PubStatus[publishData.PacketId] = protocol_stack.PUBREC
						} else if publishData.Qos == 0 {
							publishData.MQTTSeserialize_publish(&write_buf)
							client.conn.Write(write_buf)
						} else {
							log.LogPrint(log.LOG_ERROR, "[%s] error Qos level %d", routinId, publishData.Qos)
						}
					} else {
						if publishData.Qos > 0 {
							log.LogPrint(log.LOG_INFO, "[%s] client [%s] is offline, but qos(%d) > 0", routinId, client.ClientId, publishData.Qos)
							if publishData.Qos == 1 {
								client.OfflineMsg = append(client.OfflineMsg, publishData)
							} else if publishData.Qos == 2 {
								log.LogPrint(log.LOG_INFO, "[%s] client [%s] is offline msg qos(2) need wait complete", routinId, client.ClientId)
							} else {
								log.LogPrint(log.LOG_ERROR, "[%s] error Qos level %d", routinId, publishData.Qos)
							}
						}
					}
				}
			}
		}
	case MSG_REL_CLIENT:
		pub_rel_msg := msg.MsgBody.(protocol_stack.MQTTPacketPublishData)
		from := msg.MsgFrom
		for _, client := range GloablClientsMap {
			for topic := range client.SubInfo {
				if topic == pub_rel_msg.Topic {
					pub_rel_msg.PacketId = client.GeneralUniquePacketId()
					log.LogPrint(log.LOG_DEBUG, "[%s] client[%s]->client[%s] send pid %d qos2 msg", routinId, from, client.ClientId, pub_rel_msg.PacketId)
					if client.LoginSuccess == 1 {
						write_buf := []byte{}
						pub_rel_msg.MQTTSeserialize_publish(&write_buf)
						client.CacheMsg[pub_rel_msg.PacketId] = pub_rel_msg
						client.PubStatus[pub_rel_msg.PacketId] = protocol_stack.PUBREC
						client.conn.Write(write_buf)
					} else {
						log.LogPrint(log.LOG_INFO, "[%s] client [%s] msg qos2 complete but client [%s] offline need save offline", routinId, from, client.ClientId)
						client.OfflineMsg = append(client.OfflineMsg, pub_rel_msg)
					}
				}
			}
		}
	default:
		log.LogPrint(log.LOG_WARNING, "[%s] not support this msg type", routinId)
	}
}

func ClientsMapCycle(cycle chan CycleRoutinMsg) {

	var routinId string = "cycle"

	log.LogPrint(log.LOG_INFO, "[%s] start cycle", routinId)

	ticker := time.NewTicker(time.Second * 1)

	for {
		var data CycleRoutinMsg
		select {
		case <- ticker.C:
			for _, client := range GloablClientsMap {
				client.heartTimeout--
				if client.heartTimeout < 0 {
					log.LogPrint(log.LOG_WARNING, "[%s] client [%s] heart timeout", routinId, client.ClientId)
					client.conn.Close()
				}
			}
		case data = <- cycle:
			GlobalClientsMapLock.Lock()
			processMsg(routinId, data)
			GlobalClientsMapLock.Unlock()
		}
	}

}

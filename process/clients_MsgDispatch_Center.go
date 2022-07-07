package process

import (
	"mqtt_server/MQTT_Server_Go/cluster"
	"mqtt_server/MQTT_Server_Go/log"
	"mqtt_server/MQTT_Server_Go/protocol_stack"
	"time"
)

const (
	MSG_CLIENT_ADD = iota
	MSG_CLIENT_DEL
	MSG_CLIENT_SUB
	MSG_CLIENT_UNSUB
	MSG_CLIENT_PUBLISH
	MSG_CLIENT_PUBACK
	MSG_CLIENT_PUBREC
	MSG_CLIENT_PUBREL
	MSG_CLIENT_PUBCOMP

	MSG_CLUSTER_PUBLISH = cluster.CLUSTER_MSG_PUBLISH_DISPATCH
	MSG_CLUSTER_LOGIN = cluster.CLUSTER_MSG_LOGIN_NOTIFY
	MSG_CLUSTER_LOGIN_INFO = cluster.CLUSTER_MSG_LOGIN_INFO_NOTIFY
)

//type DispatchRoutinMsg struct {
//	MsgType uint16
//	MsgFrom string
//	MsgBody interface{}
//}

func processMsg(routinId string, msg cluster.RoutingCommunicateMsg) {
	switch msg.MsgType {
	case MSG_CLIENT_ADD:
		mqttClient,ok := msg.MsgBody.(*MqttClientInfo)
		if ok {
			log.LogPrint(log.LOG_DEBUG,"[%s] read from cycle channal one client login", routinId)
			GloablClientsMap[mqttClient.ClientId] = mqttClient
			var msg_cluster cluster.RoutingCommunicateMsg
			msg_cluster.MsgType = cluster.CLUSTER_LOGIN_MSG
			msg_cluster.MsgBody = mqttClient.ClientId
			cluster.Cluster_Ch <- msg_cluster
		} else {
			log.LogPrint(log.LOG_ERROR, "[%s] add client msg data is error", routinId)
		}
	case MSG_CLIENT_DEL:
		clientId,ok := msg.MsgBody.(string)
		if ok {
			delete(GloablClientsMap, clientId)
		} else {
			log.LogPrint(log.LOG_INFO, "[%s] del client msg data is error", routinId)
		}
	case MSG_CLIENT_SUB:
		subscribeData := msg.MsgBody.(protocol_stack.MQTTPacketSubscribeData)
		from := msg.MsgFrom
		self, ok := GloablClientsMap[from]
		if ok {
			subscribeData.MQTTGetSubInfo(self.SubInfo)
		}
	case MSG_CLIENT_UNSUB:
		unSubscribeData := msg.MsgBody.(protocol_stack.MQTTPacketUnSubscribeData)
		from := msg.MsgFrom
		self, ok := GloablClientsMap[from]
		if ok {
			unSubscribeData.MQTTRemoveUnSubTopic(self.SubInfo)
		}
	case MSG_CLIENT_PUBLISH:
		publishData := msg.MsgBody.(protocol_stack.MQTTPacketPublishData)
		from := msg.MsgFrom
		write_buf := []byte{}
		self, ok := GloablClientsMap[from]
		if ok {
			if publishData.Qos == 2 {
				self.PubStatus[publishData.PacketId] = protocol_stack.PUBREL
				self.CacheMsg[publishData.PacketId] = publishData
				log.LogPrint(log.LOG_INFO, "[%s] client[%s] send qos2 %d msg wait recv pub rel", routinId, from, publishData.PacketId)
			}
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
		}
	case MSG_CLIENT_PUBACK:
		publishData := msg.MsgBody.(protocol_stack.MQTTPacketPublishData)
		from := msg.MsgFrom
		self, ok := GloablClientsMap[from]
		if ok {
			status, ok := self.PubStatus[publishData.PacketId]
			if ok && status == protocol_stack.PUBACK {
				log.LogPrint(log.LOG_WARNING, "[%s] client[%s] send puback for pid %d", routinId, from, publishData.PacketId)
				delete(self.PubStatus, publishData.PacketId)
			} else {
				log.LogPrint(log.LOG_WARNING, "[%s] not find client or client status error, drop it", routinId)
			}
		}
	case MSG_CLIENT_PUBREC:
		publishData := msg.MsgBody.(protocol_stack.MQTTPacketPublishData)
		from := msg.MsgFrom
		self, ok := GloablClientsMap[from]
		if ok {
			status, ok := self.PubStatus[publishData.PacketId]
			if ok && status == protocol_stack.PUBREC {
				log.LogPrint(log.LOG_DEBUG, "[%s] client[%s] send pubrec for pid %d send pubrel qos2 msg wait pubcomp", routinId, from, publishData.PacketId)
				var write_buf_ack []byte = make([]byte, 0)
				publishData.MQTTSeserialize_pubrel(&write_buf_ack)
				self.PubStatus[publishData.PacketId] = protocol_stack.PUBCOMP
				self.conn.Write(write_buf_ack)
			} else {
				log.LogPrint(log.LOG_WARNING, "[%s] not find client or client status error, drop it", routinId)
			}
		}
	case MSG_CLIENT_PUBREL:
		publishData := msg.MsgBody.(protocol_stack.MQTTPacketPublishData)
		from := msg.MsgFrom
		self, ok := GloablClientsMap[from]
		if ok {
			var write_buf []byte = make([]byte, 0)
			status, ok := self.PubStatus[publishData.PacketId]
			if ok && status == protocol_stack.PUBREL {
				publishData.MQTTSeserialize_pubcomp(&write_buf)
				delete(self.PubStatus, publishData.PacketId)
				self.conn.Write(write_buf)
				//mqttClientData.PubStatus[publishData.PacketId] = protocol_stack.NONE
				log.LogPrint(log.LOG_DEBUG, "[%s] client[%s] send pubrel for pid %d send pubcomp qos2 msg complete", routinId, from, publishData.PacketId)
				pub_rel_msg, ok := self.CacheMsg[publishData.PacketId]
				if ok {
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
					log.LogPrint(log.LOG_INFO, "[%s] client [%s] send pubrel for pid %d msg(qos=2) complete send to cluster server node", routinId, from, publishData.PacketId)
					var msg_cluster cluster.RoutingCommunicateMsg
					msg_cluster.MsgType = cluster.CLUSTER_PUBLIC_MSG
					msg_cluster.MsgBody = pub_rel_msg
					cluster.Cluster_Ch <- msg_cluster
				} else {
					log.LogPrint(log.LOG_WARNING, "[%s] not find client msg", routinId)
				}
			} else {
				log.LogPrint(log.LOG_WARNING, "[%s] not find client or client status error, drop it", routinId)
			}
		}
	case MSG_CLIENT_PUBCOMP:
		publishData := msg.MsgBody.(protocol_stack.MQTTPacketPublishData)
		from := msg.MsgFrom
		self, ok := GloablClientsMap[from]
		status, ok := self.PubStatus[publishData.PacketId]
		if ok && status == protocol_stack.PUBCOMP {
			log.LogPrint(log.LOG_DEBUG, "[%s] client[%s] send pubcomp for pid %d qos2 msg complete delete it", routinId, from, publishData.PacketId)
			self.PubStatus[publishData.PacketId] = protocol_stack.NONE
			delete(self.CacheMsg, publishData.PacketId)
			delete(self.PubStatus, publishData.PacketId)
		} else {
			log.LogPrint(log.LOG_WARNING, "[%s] not find client or client status error, drop it", routinId)
		}
	case MSG_CLUSTER_PUBLISH:
		publishData := msg.MsgBody.(protocol_stack.MQTTPacketPublishData)
		from := msg.MsgFrom
		log.LogPrint(log.LOG_INFO, "[%s] recv from cluster server node [%s] pub msg topic %s qos %d", routinId, from, publishData.Topic, publishData.Qos)
		for _, client := range GloablClientsMap {
			for topic := range client.SubInfo {
				if topic == publishData.Topic {
					publishData.PacketId = client.GeneralUniquePacketId()
					log.LogPrint(log.LOG_DEBUG, "[%s] client[%s]->client[%s] send pid %d qos2 msg", routinId, from, client.ClientId, publishData.PacketId)
					if client.LoginSuccess == 1 {
						write_buf := []byte{}
						if publishData.Qos == 0 {
							publishData.MQTTSeserialize_publish(&write_buf)
							client.conn.Write(write_buf)
						} else if publishData.Qos == 1 {
							publishData.MQTTSeserialize_publish(&write_buf)
							client.PubStatus[publishData.PacketId] = protocol_stack.PUBACK
							client.conn.Write(write_buf)
						} else if publishData.Qos == 2 {
							publishData.MQTTSeserialize_publish(&write_buf)
							client.CacheMsg[publishData.PacketId] = publishData
							client.PubStatus[publishData.PacketId] = protocol_stack.PUBREC
							client.conn.Write(write_buf)
						} else {
							log.LogPrint(log.LOG_ERROR, "[%s] from [%s] cluster node error msg qos")
						}
					} else {
						if publishData.Qos >= 1 {
							log.LogPrint(log.LOG_INFO, "[%s] from [%s] msg qos %d complete but client [%s] offline need save offline", routinId, from, publishData.Qos, client.ClientId)
							client.OfflineMsg = append(client.OfflineMsg, publishData)
						}
					}
				}
			}
		}
	case MSG_CLUSTER_LOGIN:
		loginId := msg.MsgBody.(string)
		from := msg.MsgFrom
		find := false
		var clientInfo cluster.ClusterClientInfo
		clientInfo.Subinfo = make(map[string]uint8, 1)
		clientInfo.ClientId = loginId
		for _, client := range GloablClientsMap {
			if client.ClientId == loginId {
				log.LogPrint(log.LOG_INFO, "[%s] from [%s] cluster node client [%s] login need notify", routinId, from, loginId)
				for topic, qos := range client.SubInfo {
					clientInfo.Subinfo[topic] = qos
				}
				clientInfo.OfflineMsg = append(clientInfo.OfflineMsg, client.OfflineMsg...)
				delete(GloablClientsMap, loginId)
				find = true
				break;
			}
		}
		if find {
			var msg_cluster cluster.RoutingCommunicateMsg
			msg_cluster.MsgType = cluster.CLUSTER_CLIENT_INFO_MSG
			msg_cluster.MsgBody = clientInfo
			msg_cluster.MsgFrom = from
			cluster.Cluster_Ch <- msg_cluster
		}
	case MSG_CLUSTER_LOGIN_INFO:

	default:
		log.LogPrint(log.LOG_WARNING, "[%s] not support this msg type", routinId)
	}
}

func ClientsMsgDispatch(dispatchCh chan cluster.RoutingCommunicateMsg) {

	var routinId string = "Dispatch"

	log.LogPrint(log.LOG_INFO, "[%s] start Dispatch routing", routinId)

	ticker := time.NewTicker(time.Second * 1)

	for {
		var data cluster.RoutingCommunicateMsg
		select {
		case <- ticker.C:
			for _, client := range GloablClientsMap {
				client.heartTimeout--
				if client.heartTimeout < 0 {
					log.LogPrint(log.LOG_WARNING, "[%s] client [%s] heart timeout", routinId, client.ClientId)
					client.conn.Close()
				}
			}
		case data = <- dispatchCh:
			GlobalClientsMapLock.Lock()
			processMsg(routinId, data)
			GlobalClientsMapLock.Unlock()
		}
	}

}

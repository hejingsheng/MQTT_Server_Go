package process

import (
	"github.com/mqtt_server/MQTT_Server_Go/log"
	"github.com/mqtt_server/MQTT_Server_Go/protocol_stack"
	"net"
	"sync"
)

type MqttClientInfo struct {
	clientId string
	username string
	password string
	pingperiod uint16
	clearSession uint8
	loginSuccess byte

	//subInfo []protocol_stack.MqttSubInfo
	subInfo map[string]uint8   // subscribe topic  key = topic  value = qos
	pubStatus map[uint16]uint8  // publish msg status key = packetId value = status for qos1 and qos2 status is puback pubrec pubrel pubcomp
}

var (
	gloablClientsMap map[net.Conn]*MqttClientInfo
	globalClientsMapLock sync.Mutex
	index int = 1
)

func init() {
	gloablClientsMap = make(map[net.Conn]*MqttClientInfo, 1)
}

func ClientProcess(localconn net.Conn) {

	log.LogPrint(log.LOG_INFO, "start a tcp client")

	mqttClientData := MqttClientInfo{
		clientId:"",
		username:"",
		password:"",
		pingperiod:0,
		clearSession:0,
		loginSuccess:0,
	}

	globalClientsMapLock.Lock()
	gloablClientsMap[localconn] = &mqttClientData
	globalClientsMapLock.Unlock()

	defer func() {
		log.LogPrint(log.LOG_INFO, "one client close exit process")
		globalClientsMapLock.Lock()
		delete(gloablClientsMap, localconn)
		globalClientsMapLock.Unlock()
		localconn.Close()
	}();

	for {
		var read_buf []byte = make([]byte, 1024)
		num, err := localconn.Read(read_buf)
		if err != nil {
			log.LogPrint(log.LOG_ERROR, "read error")
			break
		} else {
			log.LogPrint(log.LOG_DEBUG, "read %d data %v", num, read_buf)
			mqttMsgType, _ := protocol_stack.MqttPacket_ParseFixHeader(read_buf)

			if mqttMsgType != -1 {
				switch mqttMsgType {
				case protocol_stack.CONNECT:
					if mqttClientData.loginSuccess == 0 {
						var conndata protocol_stack.MQTTPacketConnectData
						conndata.MQTTDeserialize_connect(read_buf, num)
						var write_buf []byte = make([]byte, 0)
						num, ret := conndata.MQTTSeserialize_connack(&write_buf)
						log.LogPrint(log.LOG_DEBUG, "write %d data, ret %d", num, ret)
						if ret == 0 {
							mqttClientData.loginSuccess = 1
							mqttClientData.pingperiod, mqttClientData.clearSession = conndata.MQTT_GetConnectInfo(&mqttClientData.username, &mqttClientData.password, &mqttClientData.clientId)
							mqttClientData.subInfo = make(map[string]uint8)
							mqttClientData.pubStatus = make(map[uint16]uint8)
						}
						localconn.Write(write_buf);
					} else {
						log.LogPrint(log.LOG_WARNING, "this client have login")
					}
				case protocol_stack.DISCONNECT:
					log.LogPrint(log.LOG_INFO, "client send disconnect req")
					localconn.Close()
				case protocol_stack.SUBSCRIBE:
					if mqttClientData.loginSuccess != 1 {
						log.LogPrint(log.LOG_ERROR, "client not login")
					} else {
						var subscribeData protocol_stack.MQTTPacketSubscribeData
						topicNum := subscribeData.MQTTDeserialize_subscribe(read_buf, num)
						log.LogPrint(log.LOG_INFO, "client subscribe %d topic", topicNum)
						var write_buf []byte = make([]byte, 0)
						subscribeData.MQTTGetSubInfo(mqttClientData.subInfo, topicNum)
						subscribeData.MQTTSeserialize_suback(&write_buf, topicNum)
						localconn.Write(write_buf);
					}
				case protocol_stack.UNSUBSCRIBE:
					if mqttClientData.loginSuccess != 1 {
						log.LogPrint(log.LOG_ERROR, "client not login")
					} else {
						var unSubscribeData protocol_stack.MQTTPacketUnSubscribeData
						topicNum := unSubscribeData.MQTTDeserialize_unsubscribe(read_buf, num)
						log.LogPrint(log.LOG_INFO, "client unsubscribe %d topic", topicNum)
						var write_buf []byte = make([]byte, 0)
						unSubscribeData.MQTTRemoveUnSubTopic(mqttClientData.subInfo)
						unSubscribeData.MQTTSeserialize_unsuback(&write_buf)
						localconn.Write(write_buf);
					}
				case protocol_stack.PUBLISH:
					if mqttClientData.loginSuccess != 1 {
						log.LogPrint(log.LOG_ERROR, "client not login")
					} else {
						var publishData protocol_stack.MQTTPacketPublishData
						var pubInfo protocol_stack.MQTTPubInfo
						ret := publishData.MQTTDeserialize_publish(read_buf, num)
						publishData.MQTTGetPublishInfo(&pubInfo)
						log.LogPrint(log.LOG_INFO, "client public msg %d, public %s,payload %d", ret, pubInfo.Topic, pubInfo.Qos)
						var write_buf []byte = read_buf[:num]
						for conn, client := range gloablClientsMap {
							//if conn == localconn{
							//	continue
							//}
							for topic := range client.subInfo {
								if topic == pubInfo.Topic {
									conn.Write(write_buf)
									if pubInfo.Qos == 1 {
										mqttClientData.pubStatus[pubInfo.PacketId] = protocol_stack.PUBACK
									} else if pubInfo.Qos == 2 {
										//mqttClientData.pubStatus[pubInfo.PacketId] = protocol_stack.PUBREC
									} else if pubInfo.Qos == 0 {

									} else {
										log.LogPrint(log.LOG_ERROR, "error Qos level %d", pubInfo.Qos)
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
					for conn, client := range gloablClientsMap {
						//if conn == localconn{
						//	continue
						//}
						status, ok := client.pubStatus[pubInfo.PacketId]
						if ok && status == protocol_stack.PUBACK {
							conn.Write(write_buf)
							delete(client.pubStatus, pubInfo.PacketId)
						} else {
							log.LogPrint(log.LOG_WARNING, "not find client or client status error, drop it")
						}
					}
				case protocol_stack.PUBREC:
				case protocol_stack.PUBREL:
				case protocol_stack.PUBCOMP:

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
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
	pingperiod int
	clearSession int
	loginSuccess byte
}

var (
	gloablClientsMap map[net.Conn]MqttClientInfo
	globalClientsMapLock sync.Mutex
	index int = 1
)

func init() {
	gloablClientsMap = make(map[net.Conn]MqttClientInfo, 1)
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
	gloablClientsMap[localconn] = mqttClientData
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
						}
						localconn.Write(write_buf);
					} else {
						log.LogPrint(log.LOG_WARNING, "this client have login")
					}
				case protocol_stack.DISCONNECT:
					log.LogPrint(log.LOG_INFO, "client send disconnect req")
					localconn.Close()
				case protocol_stack.PINGREQ:
					var pingpongdata protocol_stack.MQTTPacketPingPongData
					pingpongdata.MQTTDeserialize_ping(read_buf, num)
					var write_buf []byte = make([]byte, 0)
					num = pingpongdata.MQTTSeserialize_pong(&write_buf)
					localconn.Write(write_buf)
				}
			}

			for conn, _ := range gloablClientsMap {
				if conn == localconn{
					continue
				}
				//fmt.Println("this is ", index, " conn")
				//conn.Write(write_buf);
			}
		}
	}
}
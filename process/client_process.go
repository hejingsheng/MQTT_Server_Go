package process

import (
	"fmt"
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
	//ret := C.add(1,2)
	//fmt.Println("ret=",ret)
	fmt.Println("start a tcp client")

	mqttClientData := MqttClientInfo{
		clientId:"",
		username:"",
		password:"",
		pingperiod:0,
		clearSession:0,
	}

	globalClientsMapLock.Lock()
	gloablClientsMap[localconn] = mqttClientData
	globalClientsMapLock.Unlock()

	defer func() {
		fmt.Println("one client close exit process")
		globalClientsMapLock.Lock()
		delete(gloablClientsMap, localconn)
		globalClientsMapLock.Unlock()
		localconn.Close()
	}();

	for {
		var read_buf []byte = make([]byte, 1024)
		num, err := localconn.Read(read_buf)
		if err != nil {
			fmt.Println("read error")
			break
		} else {
			fmt.Printf("read %d data %s\n", num, read_buf)
			var conndata protocol_stack.MQTTPacketConnectData
			conndata.MQTTDeserialize_connect(read_buf, num)
			var write_buf []byte = make([]byte, 0)
			num = conndata.MQTTSeserialize_connack(&write_buf)
			fmt.Printf("write %d data\n", num)
			for conn, index := range gloablClientsMap {
				if conn == localconn{
					continue
				}
				fmt.Println("this is ", index, " conn")
				conn.Write(write_buf);
			}
		}
	}
}
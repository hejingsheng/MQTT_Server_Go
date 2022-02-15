package protocol_stack

import (
	"fmt"
)

const (
	NONE = iota
	CONNECT
	CONNACK
	PUBLISH
	PUBACK
	PUBREC
	PUBREL
	PUBCOMP
	SUBSCRIBE
	SUBACK
	UNSUBSCRIBE
	UNSUBACK
	PINGREQ
	PINGRESP
	DISCONNECT
)

type MQTTPacketWillOptions struct {
	topic string
	message string
	retained byte
	qos int
}

type MQTTPacketConnectData struct {
	version int
    clientID string
	keepAliveInterval int
    cleansession int
	wills MQTTPacketWillOptions
    username string
	password string
}

func (connData *MQTTPacketConnectData) MQTTDeserialize_connect(data []byte, len int) int {
	var value int
	var version int
	var connectFlag byte

	index := 0
	firstByte := data[index]
	index++
	mqttDataType := (firstByte & 0xF0) >> 4

	if mqttDataType != CONNECT {
		return -1
	}

	leftdata := data[index:len]
	index += mqttPacket_decode(leftdata, &value) // read remaining length
	fmt.Println("index=",index," value=",value)
	leftdata = data[index:len]
	index += mqttPacket_readProtocol(leftdata, &version) //read protocol name
	fmt.Println("index=",index," version=",version)
	connData.version = version
	connectFlag = data[index]
	index++
	connData.cleansession = int(connectFlag & 0x02 >> 1)
	connData.keepAliveInterval = int(data[index] << 8 | data[index+1])
	index += 2
	leftdata = data[index:len]
	index += mqttPacket_readString(leftdata, &connData.clientID)
	if connectFlag & 0x04 != 0 {
		connData.wills.qos = int(connectFlag & 0x18 >> 3)
		connData.wills.retained = byte(connectFlag & 0x20 >> 5)
		leftdata = data[index:len]
		index += mqttPacket_readString(leftdata, &connData.wills.topic)
		leftdata = data[index:len]
		index += mqttPacket_readString(leftdata, &connData.wills.message)
	}
	if connectFlag & 0x80 != 0 {
		leftdata = data[index:len]
		index += mqttPacket_readString(leftdata, &connData.username)
		leftdata = data[index:len]
		index += mqttPacket_readString(leftdata, &connData.password)
	}
	return 0
}

func (connData *MQTTPacketConnectData) MQTTSeserialize_connack(data *[]byte) (int, byte) {
	var header byte
	var retCode byte
	index := 0
	header = CONNACK
	header <<= 4
	*data = append(*data, header)
	index++
	tmp, leftLen := mqttPacket_encode(2);
	for i := 0; i < tmp; i++ {
		*data = append(*data, leftLen[i])
	}
	index += tmp
	*data = append(*data, 0x01)
	index++
	if connData.username == connData.password {
		retCode = 0x00
	} else {
		retCode = 0x04
	}
	*data = append(*data, retCode)
	index++
	return index, retCode
}

func (connData *MQTTPacketConnectData) MQTT_GetConnectInfo(name *string, pass *string, clientId *string) (int, int) {
	*name = connData.username
	*pass = connData.password
	*clientId = connData.clientID
	return connData.keepAliveInterval, connData.cleansession
}
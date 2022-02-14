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

func mqttPacket_decode(data []byte, value *int) int {
	var c byte
	var multiplier int = 1
	var len int = 0
	i := 0

	for {
		len++
		if len > 4 {
			return 0
		}
		c = data[i]
		*value += int(c & 127) * multiplier
		multiplier *= 128
		if data[i] & 128 == 0 {
			break;
		}
		i++
	}
	return len
}

func mqttPacket_encode(length int) (int, [4]byte) {
	var buf [4]byte
	var index int = 0
	for {
		var tmp byte
		tmp = byte(length % 128)
		length /= 128
		if length > 0 {
			tmp |= 0x80
		}
		buf[index] = tmp;
		index++
		if length <= 0 {
			break
		}
	}
	return index, buf
}

func mqttPacket_readString(data []byte, str *string) int {
	var len int = 0
	len = (int)(data[0] << 8 | data[1])
	var tmp []byte = make([]byte, 0)
	for i := 0; i < len; i++  {
		tmp = append(tmp, data[i+2])
	}
	*str = string(tmp)
	return len+2
}

func mqttPacket_readProtocol(data []byte, version *int) int {
	var len int
	len = (int)(data[0] << 8 | data[1])
	var tmp []byte = make([]byte, 0)
	for i := 0; i < len; i++  {
		tmp = append(tmp, data[i+2])
	}
	str := string(tmp)
	fmt.Println("read str=",str)
	*version = int(data[2+len])
	return 2+len+1
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
	//read protocol name
	leftdata = data[index:len]
	index += mqttPacket_readProtocol(leftdata, &version)
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

func (connData *MQTTPacketConnectData) MQTTSeserialize_connack(data *[]byte) int {
	var header byte
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
		*data = append(*data, 0x00)
		index++
	}
	return index
}
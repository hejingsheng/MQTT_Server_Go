package protocol_stack

const (
	CONNECT = 1
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

type MQTTPacketConnectData struct {
	version int
    clientID string
	keepAliveInterval int
    cleansession int
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

func (connData *MQTTPacketConnectData) MQTTDeserialize_connect(data []byte, len int) int {

	var value int
	firstByte := data[0]
	mqttDataType := firstByte & 0xF0

	if mqttDataType != CONNECT {
		//return -1
	}

	leftdata := data[1:len]
	mqttPacket_decode(leftdata, &value)


	return 0
}

package protocol_stack

type MqttSubInfo struct {
	Topic string
	Qos   uint8
}

type MQTTPacketSubscribeData struct {
	Dup      uint8
	PacketId uint16
	Topic    []string
	Qos      []uint8
}

func (subdata *MQTTPacketSubscribeData) MQTTDeserialize_subscribe(buf []byte, len int) int {
	var header byte = 0
	var remainLen int = 0
	var topicNum int = 0
	index := 0

	header = buf[index]
	index++
	mqttDataType := int(header & 0xF0 >> 4)

	if mqttDataType != SUBSCRIBE {
		return -1
	}
	subdata.Dup = uint8(header & 0x08 >> 3)
	leftdata := buf[index:len]
	index += mqttPacket_decode(leftdata, &remainLen) // read remaining length
	subdata.PacketId = uint16(buf[index])<<8 | uint16(buf[index+1])
	index += 2
	for {
		if index >= len {
			break
		}
		var topic string
		var qos uint8
		leftdata = buf[index:len]
		index += mqttPacket_readString(leftdata, &topic)
		subdata.Topic = append(subdata.Topic, topic)
		qos = buf[index] & 0x03
		subdata.Qos = append(subdata.Qos, qos)
		index++
		topicNum++
	}
	return topicNum
}

func (subdata *MQTTPacketSubscribeData) MQTTSeserialize_suback(buf *[]byte, num int) int {
	var header byte
	index := 0
	header = SUBACK
	header <<= 4

	*buf = append(*buf, header)
	index++
	tmp, leftLen := mqttPacket_encode(2 + num)
	for i := 0; i < tmp; i++ {
		*buf = append(*buf, leftLen[i])
	}
	index += tmp
	*buf = append(*buf, uint8(subdata.PacketId/256))
	*buf = append(*buf, uint8(subdata.PacketId%256))
	index += 2
	for i := 0; i < num; i++ {
		*buf = append(*buf, subdata.Qos[i])
		index++
	}
	return index
}

func (subdata *MQTTPacketSubscribeData) MQTTGetSubInfo(subInfo map[string]uint8) {
	for i := 0; i < len(subdata.Topic); i++ {
		subInfo[subdata.Topic[i]] = subdata.Qos[i]
	}
}

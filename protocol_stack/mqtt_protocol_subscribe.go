package protocol_stack

type MQTTPacketSubscribeData struct {
	dup uint8
	packetId uint16
	topic []string
	qos []uint8
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
	subdata.dup = uint8(header & 0x08 >> 3)
	leftdata := buf[index:len]
	index += mqttPacket_decode(leftdata, &remainLen) // read remaining length
	subdata.packetId = uint16(buf[index]) << 8 | uint16(buf[index+1])
	index += 2
	for {
		if index >= len {
			break;
		}
		var topic string
		var qos uint8
		leftdata = buf[index:len]
		index += mqttPacket_readString(leftdata, &topic)
		subdata.topic = append(subdata.topic, topic)
		qos = buf[index] & 0x03
		subdata.qos = append(subdata.qos, qos)
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
	tmp, leftLen := mqttPacket_encode(2+num)
	for i := 0; i < tmp; i++ {
		*buf = append(*buf, leftLen[i])
	}
	index += tmp
	*buf = append(*buf, uint8(subdata.packetId/256))
	*buf = append(*buf, uint8(subdata.packetId%256))
	index += 2
	for i := 0; i < num; i++ {
		*buf = append(*buf, subdata.qos[i])
		index++
	}
	return index
}
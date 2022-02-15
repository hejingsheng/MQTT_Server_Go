package protocol_stack

type MQTTPacketUnSubscribeData struct {
	topic []string
	packetId uint16
}

func (unsubData *MQTTPacketUnSubscribeData)MQTTDeserialize_unsubscribe(buf []byte, len int) int {
	var header byte = 0
	var remainLen int = 0
	var topicNum int = 0
	index := 0

	header = buf[index]
	index++
	mqttDataType := int(header & 0xF0 >> 4)

	if mqttDataType != UNSUBSCRIBE {
		return -1
	}
	if header & 0x0f != 0x02 {  // UNSUBSCRIBE报文固定报头的第3,2,1,0位是保留位且必须分别设置为0,0,1,0
		return -1
	}
	leftdata := buf[index:len]
	index += mqttPacket_decode(leftdata, &remainLen)
	unsubData.packetId = uint16(buf[index]) << 8 | uint16(buf[index+1])
	index += 2
	for {
		if index >= len {
			break
		}
		var topic string
		leftdata = buf[index:len]
		index += mqttPacket_readString(leftdata, &topic)
		topicNum++
		unsubData.topic = append(unsubData.topic, topic)
	}
	return topicNum
}

func (unsubData *MQTTPacketUnSubscribeData)MQTTSeserialize_unsuback(buf *[]byte) int {
	var header byte
	index := 0

	header = UNSUBACK
	header <<= 4
	*buf = append(*buf, header)
	index++
	tmp, leftLen := mqttPacket_encode(2)
	for i := 0; i < tmp; i++ {
		*buf = append(*buf, leftLen[i])
	}
	index += tmp
	*buf = append(*buf, uint8(unsubData.packetId/256))
	*buf = append(*buf, uint8(unsubData.packetId%256))
	index += 2
	return index
}
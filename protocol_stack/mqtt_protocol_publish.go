package protocol_stack

//type MQTTPubInfo struct {
//	Topic string
//	Qos uint8
//	Payload string
//	PacketId uint16
//}

type MQTTPacketPublishData struct {
	Dup uint8        `json:"dup"`
	PacketId uint16  `json:"packetid"`
	Retained uint8   `json:"retained"`
	Topic string     `json:"topic"`
	Qos uint8        `json:"qos"`
	Payload string   `json:"payload"`
}

func (publishData *MQTTPacketPublishData)MQTTDeserialize_publish(buf []byte, len int) int {
	var header byte = 0
	var remainLen int = 0
	index := 0

	header = buf[index]
	index++
	mqttDataType := int(header & 0xF0 >> 4)
	if mqttDataType != PUBLISH {
		return -1;
	}
	publishData.Dup = header & 0x08 >> 3
	publishData.Qos = header & 0x06 >> 1
	publishData.Retained = header & 0x01
	leftdata := buf[index:len]
	index += mqttPacket_decode(leftdata, &remainLen)
	leftdata = buf[index:len]
	index += mqttPacket_readString(leftdata, &publishData.Topic)
	if publishData.Qos > 0 {
		publishData.PacketId = uint16(buf[index]) << 8 | uint16(buf[index+1])
		index += 2
	}
	leftdata = buf[index:len]
	publishData.Payload = string(leftdata)
	return 0
}

func (publishData *MQTTPacketPublishData)MQTTSeserialize_publish(buf *[]byte) int {
	var header byte = 0
	var remainLen int = 0
	index := 0

	header = PUBLISH
	header <<= 4
	header |= publishData.Qos << 1
	header |= publishData.Dup << 3
	header |= publishData.Retained
	*buf = append(*buf, header)
	index++
	remainLen = 2+len(publishData.Topic)+len(publishData.Payload)
	if publishData.Qos > 0 {
		remainLen += 2
	}
	tmp, leftLen := mqttPacket_encode(remainLen)
	for i := 0; i < tmp; i++ {
		*buf = append(*buf, leftLen[i])
	}
	index += tmp
	*buf = append(*buf, uint8(len(publishData.Topic)/256))
	*buf = append(*buf, uint8(len(publishData.Topic)%256))
	index += 2
	t := []byte(publishData.Topic)
	*buf = append(*buf, t...)
	index += len(publishData.Topic)
	if publishData.Qos > 0 {
		*buf = append(*buf, uint8(publishData.PacketId/256))
		*buf = append(*buf, uint8(publishData.PacketId%256))
		index += 2
	}
	t = []byte(publishData.Payload)
	*buf = append(*buf, t...)
	index += len(publishData.Payload)
	return index
}

func MQTTDeserialize_ack(buf []byte, len int, msgType uint8, packetId *uint16) int {
	var header byte
	var remainLen int
	index := 0
	header = buf[index]
	index++
	mqttDataType := uint8(header & 0xF0 >> 4)

	if mqttDataType != msgType {
		return -1
	}
	leftData := buf[index:len]
	index += mqttPacket_decode(leftData, &remainLen)
	*packetId = uint16(buf[index]) << 8 | uint16(buf[index+1])
	index += 2
	return index
}

func (publishData *MQTTPacketPublishData)MQTTDeserialize_puback(buf []byte, len int) int {
	return MQTTDeserialize_ack(buf, len, PUBACK, &publishData.PacketId)
}

func (publishData *MQTTPacketPublishData)MQTTDeserialize_pubrec(buf []byte, len int) int {
	return MQTTDeserialize_ack(buf, len, PUBREC, &publishData.PacketId)
}

func (publishData *MQTTPacketPublishData)MQTTDeserialize_pubrel(buf []byte, len int) int {
	return MQTTDeserialize_ack(buf, len, PUBREL, &publishData.PacketId)
}

func (publishData *MQTTPacketPublishData)MQTTDeserialize_pubcomp(buf []byte, len int) int {
	return MQTTDeserialize_ack(buf, len, PUBCOMP, &publishData.PacketId)
}

func MQTTSeserialize_ack(buf *[]byte, msgType uint8, publishData *MQTTPacketPublishData) int {
	var header byte = 0
	index := 0

	header = msgType
	header <<= 4
	if msgType == PUBREL {
		header |= 1 << 1
	} else {
		header |= 0 << 1
	}
	header |= publishData.Dup << 3
	header |= publishData.Retained
	*buf = append(*buf, header)
	index++
	tmp, leftLen := mqttPacket_encode(2)
	for i := 0; i < tmp; i++ {
		*buf = append(*buf, leftLen[i])
	}
	index += tmp
	*buf = append(*buf, uint8(publishData.PacketId/256))
	*buf = append(*buf, uint8(publishData.PacketId%256))
	index += 2
	return index
}

func (publishData *MQTTPacketPublishData)MQTTSeserialize_puback(buf *[]byte) int {
	return MQTTSeserialize_ack(buf, PUBACK, publishData)
}

func (publishData *MQTTPacketPublishData)MQTTSeserialize_pubrec(buf *[]byte) int {
	return MQTTSeserialize_ack(buf, PUBREC, publishData)
}

func (publishData *MQTTPacketPublishData)MQTTSeserialize_pubrel(buf *[]byte) int {
	return MQTTSeserialize_ack(buf, PUBREL, publishData)
}

func (publishData *MQTTPacketPublishData)MQTTSeserialize_pubcomp(buf *[]byte) int {
	return MQTTSeserialize_ack(buf, PUBCOMP, publishData)
}

//func (publishData *MQTTPacketPublishData)MQTTGetPublishInfo(pubInfo *MQTTPubInfo) int {
//	pubInfo.Topic = publishData.topic
//	pubInfo.Qos = publishData.qos
//	pubInfo.Payload = publishData.payload
//	pubInfo.PacketId = publishData.packetId
//	return 0
//}
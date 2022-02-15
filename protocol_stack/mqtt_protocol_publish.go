package protocol_stack

type MQTTPubInfo struct {
	Topic string
	Qos uint8
	Payload string
}

type MQTTPacketPublishData struct {
	dup uint8
	packetId uint16
	retained uint8
	topic string
	qos uint8
	payload string
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
	publishData.dup = header & 0x08 >> 3
	publishData.qos = header & 0x06 >> 1
	publishData.retained = header & 0x01
	leftdata := buf[index:len]
	index += mqttPacket_decode(leftdata, &remainLen)
	leftdata = buf[index:len]
	index += mqttPacket_readString(leftdata, &publishData.topic)
	if publishData.qos > 0 {
		publishData.packetId = uint16(buf[index]) << 8 | uint16(buf[index+1])
		index += 2
	}
	leftdata = buf[index:len]
	publishData.payload = string(leftdata)
	return 0
}

//func MQTTDeserialize_ack(buf []byte, msgType *uint8, packetId *uint16) int {
//	var header byte
//	index := 0
//	header = msgType
//	header <<= 4
//	*buf = append(*buf, header)
//	index++
//	tmp, leftLen := mqttPacket_encode(2)
//	for i := 0; i < tmp; i++ {
//		*buf = append(*buf, leftLen[i])
//	}
//	index += tmp
//	*buf = append(*buf, uint8(packetId/256))
//	*buf = append(*buf, uint8(packetId%256))
//	index += 2
//	return index
//}
//
//func (publishData *MQTTPacketPublishData)MQTTSeserialize_puback(buf *[]byte, msgType uint8) int {
//	return MQTTSeserialize_puback(buf, PUBACK, publishData.packetId)
//}
//
//func (publishData *MQTTPacketPublishData)MQTTSeserialize_pubrec(buf *[]byte, msgType uint8) int {
//	return MQTTSeserialize_puback(buf, PUBREC, publishData.packetId)
//}
//
//func (publishData *MQTTPacketPublishData)MQTTSeserialize_pubrel(buf *[]byte, msgType uint8) int {
//	return MQTTSeserialize_puback(buf, PUBREL, publishData.packetId)
//}
//
//func (publishData *MQTTPacketPublishData)MQTTSeserialize_pubcomp(buf *[]byte, msgType uint8) int {
//	return MQTTSeserialize_puback(buf, PUBCOMP, publishData.packetId)
//}

func (publishData *MQTTPacketPublishData)MQTTGetPublishInfo(pubInfo *MQTTPubInfo) int {
	pubInfo.Topic = publishData.topic
	pubInfo.Qos = publishData.qos
	pubInfo.Payload = publishData.payload
	return 0
}
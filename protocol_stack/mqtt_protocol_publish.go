package protocol_stack

type MQTTPubInfo struct {
	Topic string
	Qos uint8
	Payload string
	PacketId uint16
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
	return MQTTDeserialize_ack(buf, len, PUBACK, &publishData.packetId)
}

func (publishData *MQTTPacketPublishData)MQTTDeserialize_pubrec(buf []byte, len int) int {
	return MQTTDeserialize_ack(buf, len, PUBREC, &publishData.packetId)
}

func (publishData *MQTTPacketPublishData)MQTTDeserialize_pubrel(buf []byte, len int) int {
	return MQTTDeserialize_ack(buf, len, PUBREL, &publishData.packetId)
}

func (publishData *MQTTPacketPublishData)MQTTDeserialize_pubcomp(buf []byte, len int) int {
	return MQTTDeserialize_ack(buf, len, PUBCOMP, &publishData.packetId)
}

func (publishData *MQTTPacketPublishData)MQTTGetPublishInfo(pubInfo *MQTTPubInfo) int {
	pubInfo.Topic = publishData.topic
	pubInfo.Qos = publishData.qos
	pubInfo.Payload = publishData.payload
	pubInfo.PacketId = publishData.packetId
	return 0
}
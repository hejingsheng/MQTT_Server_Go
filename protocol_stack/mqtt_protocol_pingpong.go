package protocol_stack

type MQTTPacketPingPongData struct {

}

func (ping *MQTTPacketPingPongData)MQTTDeserialize_ping(data []byte, len int) int {
	index := 0
	firstByte := data[index]
	index++
	mqttDataType := (firstByte & 0xF0) >> 4

	if mqttDataType != PINGREQ {
		return -1
	}
	return 0
}

func (pong *MQTTPacketPingPongData)MQTTSeserialize_pong(data *[]byte) int {
	var header byte
	index := 0
	header = PINGRESP
	header <<= 4
	*data = append(*data, header)
	index++
	tmp, leftLen := mqttPacket_encode(0);
	for i := 0; i < tmp; i++ {
		*data = append(*data, leftLen[i])
	}
	index += tmp
	return index
}
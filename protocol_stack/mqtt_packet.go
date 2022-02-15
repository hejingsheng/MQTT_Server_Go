package protocol_stack

import _"fmt"

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
	*version = int(data[2+len])
	return 2+len+1
}

func MqttPacket_ParseFixHeader(buf []byte) (int, int) {
	var msgType int = -1
	var length int = 0
	header := buf[0]
	msgType = int(header & 0xF0 >> 4)
	if msgType > DISCONNECT {
		//msg type error
		return -1,0
	}
 	leftData := buf[1:]
	mqttPacket_decode(leftData, &length)
	return msgType,length
}
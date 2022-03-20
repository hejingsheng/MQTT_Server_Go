package utils

import (
	"math/rand"
	"time"
)

var (
	asciiCode string = "0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	RoutinIdMap map[string]uint8
)

func init() {
	rand.Seed(time.Now().Unix())
	RoutinIdMap = make(map[string]uint8, 0)
}

func generalRandNum() int {
	return rand.Intn(len(asciiCode))
}

func GeneralRoutinId(len int) string {

	var id []byte = make([]byte, len)

	for   {
		for i := 0; i < len; i++  {
			index := generalRandNum()
			id[i] = asciiCode[index]
		}
		tmp := string(id)
		_, ok := RoutinIdMap[tmp]
		if !ok {
			RoutinIdMap[tmp] = 0
			return tmp
		}
	}
}

func RemoveRoutinId(id string) {
	_, ok := RoutinIdMap[id]
	if ok {
		delete(RoutinIdMap, id)
	}
}

func GeneralPacketID() uint16 {
	id := uint16(generalRandNum()&0x0000ffff)
	return id
}



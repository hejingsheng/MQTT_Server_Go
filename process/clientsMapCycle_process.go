package process

import (
	"github.com/mqtt_server/MQTT_Server_Go/log"
	"time"
)

func ClientsMapCycle(cycle chan int) {

	var routinId string = "cycle"

	log.LogPrint(log.LOG_INFO, routinId, "start cycle")

	ticker := time.NewTicker(time.Second * 1)

	for {
		select {
		case <- ticker.C:
			//log.LogPrint(log.LOG_DEBUG, routinId,"timer coming!!!")
		case <- cycle:
			log.LogPrint(log.LOG_DEBUG, routinId,"read from cycle channal one client login")

		}
	}

}

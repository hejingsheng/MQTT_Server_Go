package process

import (
	"mqtt_server/MQTT_Server_Go/log"
	"time"
)

const (
	MSG_ADD_CLIENT = iota
	MSG_DEL_CLIENT
)

type CycleRoutinMsg struct {
	MsgType uint16
	MsgBody interface{}
}

func processMsg(routinId string, msg CycleRoutinMsg) {
	switch msg.MsgType {
	case MSG_ADD_CLIENT:
		mqttClient,ok := msg.MsgBody.(*MqttClientInfo)
		if ok {
			log.LogPrint(log.LOG_DEBUG, routinId,"read from cycle channal one client login")
			GloablClientsMap[mqttClient.ClientId] = mqttClient
		} else {
			log.LogPrint(log.LOG_ERROR, routinId, "add client msg data is error")
		}
	case MSG_DEL_CLIENT:
		clientId,ok := msg.MsgBody.(string)
		if ok {
			delete(GloablClientsMap, clientId)
		} else {
			log.LogPrint(log.LOG_INFO, routinId, "del client msg data is error")
		}
	default:
		log.LogPrint(log.LOG_WARNING, routinId, "not support this msg type")
	}
}

func ClientsMapCycle(cycle chan CycleRoutinMsg) {

	var routinId string = "cycle"

	log.LogPrint(log.LOG_INFO, routinId, "start cycle")

	ticker := time.NewTicker(time.Second * 1)

	for {
		var data CycleRoutinMsg
		select {
		case <- ticker.C:
			for _, client := range GloablClientsMap {
				client.heartTimeout--
				if client.heartTimeout < 0 {
					log.LogPrint(log.LOG_WARNING, routinId, "client [%s] heart timeout", client.ClientId)
					client.conn.Close()
				}
			}
		case data = <- cycle:
			processMsg(routinId, data)
		}
	}

}

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
			log.LogPrint(log.LOG_DEBUG,"[%s] read from cycle channal one client login", routinId)
			GloablClientsMap[mqttClient.ClientId] = mqttClient
		} else {
			log.LogPrint(log.LOG_ERROR, "[%s] add client msg data is error", routinId)
		}
	case MSG_DEL_CLIENT:
		clientId,ok := msg.MsgBody.(string)
		if ok {
			delete(GloablClientsMap, clientId)
		} else {
			log.LogPrint(log.LOG_INFO, "[%s] del client msg data is error", routinId)
		}
	default:
		log.LogPrint(log.LOG_WARNING, "[%s] not support this msg type", routinId)
	}
}

func ClientsMapCycle(cycle chan CycleRoutinMsg) {

	var routinId string = "cycle"

	log.LogPrint(log.LOG_INFO, "[%s] start cycle", routinId)

	ticker := time.NewTicker(time.Second * 1)

	for {
		var data CycleRoutinMsg
		select {
		case <- ticker.C:
			for _, client := range GloablClientsMap {
				client.heartTimeout--
				if client.heartTimeout < 0 {
					log.LogPrint(log.LOG_WARNING, "[%s] client [%s] heart timeout", routinId, client.ClientId)
					client.conn.Close()
				}
			}
		case data = <- cycle:
			processMsg(routinId, data)
		}
	}

}

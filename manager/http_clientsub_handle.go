package manager

import (
	"encoding/json"
	"fmt"
	"mqtt_server/MQTT_Server_Go/process"
	"net/http"
	"strconv"
)

func GetClientSubList(clientId string, start int, limit int) []SubInfo  {
	client, ok := process.GloablClientsMap[clientId]
	if ok {
		var data []SubInfo = make([]SubInfo, 0)
		var index int = 0
		for topic, qos := range client.SubInfo {
			if index >= start && index < start+limit {
				var tmp SubInfo
				tmp.Topic = topic
				tmp.Qos = qos
				data = append(data, tmp)
			}
			index++
		}
		return data
	} else {
		return nil
	}
}

type SubInfo struct {
	Topic        string `json:"topic"`
	Qos          uint8  `json:"qos"`
}

type ClientSubInfoHandle struct {
	ClientSubInfoList []SubInfo `json:"clientsubinfolist"`
}

func (list *ClientSubInfoHandle)ServeHTTP(w http.ResponseWriter, r *http.Request) {

	arg1 := r.URL.Query().Get("start")
	arg2 := r.URL.Query().Get("limit")
	arg3 := r.URL.Query().Get("clientId")
	start, err := strconv.Atoi(arg1)
	if err != nil {
		return
	}
	limit, err := strconv.Atoi(arg2)
	if err != nil {
		return
	}
	list.ClientSubInfoList = GetClientSubList(arg3, start, limit)
	jsonResp, err := json.Marshal(list)
	if err != nil {
		fmt.Println(err)
		return
	}
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Content-Type", "application/json")
	w.Write(jsonResp)
}
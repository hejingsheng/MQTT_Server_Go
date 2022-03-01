package manager

import (
	"encoding/json"
	"fmt"
	"github.com/mqtt_server/MQTT_Server_Go/process"
	_"github.com/mqtt_server/MQTT_Server_Go/process"
	"net/http"
)

func GetClientInfoList(start int, limit int) []ClientInfo {
	var data []ClientInfo = make([]ClientInfo, limit)
	var index int = 0
	for _, client := range process.GloablClientsMap {
		if index >= start && index < start+limit {
			var tmp ClientInfo
			tmp.ClientId = client.ClientId
			tmp.LoginSuccess = client.LoginSuccess
			tmp.Username = client.Username
			tmp.SubTopicNum = len(client.SubInfo)
			tmp.OfflineNum = len(client.OfflineMsg)
			data[index-start] = tmp
		}
		index++
	}
	return data
}

type ClientInfo struct {
	ClientId     string `json:"clientid"`
	Username     string `json:"username"`
	LoginSuccess byte   `json:"loginsuccess"`
	SubTopicNum  int    `json:"subtopicnum"`
	OfflineNum   int    `json:"offlinemsgnum"`
}

type ClientInfoListHandle struct {
	ClientInfoList []ClientInfo `json:"clientinfolist"`
}

func (list *ClientInfoListHandle)ServeHTTP(w http.ResponseWriter, r *http.Request) {

	list.ClientInfoList = GetClientInfoList(0, 10)
	jsonResp, err := json.Marshal(list)
	if err != nil {
		fmt.Println(err)
		return
	}
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Content-Type", "application/json")
	w.Write(jsonResp)
}
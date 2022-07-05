package manager

import (
	"encoding/json"
	"fmt"
	"mqtt_server/MQTT_Server_Go/process"
	"net/http"
	"strconv"
)

func GetClientInfoList(start int, limit int) []ClientInfo {
	var data []ClientInfo = make([]ClientInfo, 0)
	var index int = 0
	var length int = 0
	process.GlobalClientsMapLock.RLock()
	for _, client := range process.GloablClientsMap {
		if index >= start && index < start+limit {
			var tmp ClientInfo
			tmp.ClientId = client.ClientId
			tmp.LoginSuccess = client.LoginSuccess
			tmp.Username = client.Username
			tmp.SubTopicNum = len(client.SubInfo)
			tmp.SubTopicNum = length
			tmp.OfflineNum = len(client.OfflineMsg)
			data = append(data, tmp)
		}
		index++
	}
	process.GlobalClientsMapLock.RUnlock()
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

	arg1 := r.URL.Query().Get("start")
	arg2 := r.URL.Query().Get("limit")
	start, err := strconv.Atoi(arg1)
	if err != nil {
		return
	}
	limit, err := strconv.Atoi(arg2)
	if err != nil {
		return
	}
	list.ClientInfoList = GetClientInfoList(start, limit)
	jsonResp, err := json.Marshal(list)
	if err != nil {
		fmt.Println(err)
		return
	}
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Content-Type", "application/json")
	w.Write(jsonResp)
}
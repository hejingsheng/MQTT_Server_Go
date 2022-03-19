package manager

import (
	"encoding/json"
	"fmt"
	"mqtt_server/MQTT_Server_Go/process"
	"net/http"
)

func GetOnLineClientNum() int {
	var num int = 0
	for _, client := range process.GloablClientsMap {
		if client.LoginSuccess == 1 {
			num++
		}
	}
	return num
}

func GetSessionNum() int {
	return len(process.GloablClientsMap)
}

type OnLineNumHandle struct {
	Num  int  `json:"num"`
}

func (online *OnLineNumHandle)ServeHTTP(w http.ResponseWriter, r *http.Request) {
	online.Num = GetOnLineClientNum()
	jsonResp, err := json.Marshal(online)
	if err != nil {
		fmt.Println(err)
		return
	}
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Content-Type", "application/json")
	w.Write(jsonResp)
}

type SessionNumHandel struct {
	Num  int  `json:"num"`
}

func (session *SessionNumHandel)ServeHTTP(w http.ResponseWriter, r *http.Request) {
	session.Num = GetSessionNum()
	jsonResp, err := json.Marshal(session)
	if err != nil {
		fmt.Println(err)
		return
	}
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Content-Type", "application/json")
	w.Write(jsonResp)
}
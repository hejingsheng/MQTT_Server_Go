package manager

import (
	"encoding/json"
	"fmt"
	"github.com/mqtt_server/MQTT_Server_Go/process"
	"net/http"
)

func GetClientOffLineMsgList(clientId string) int  {
	client, ok := process.GloablClientsMap[clientId]
	if ok {
		return len(client.OfflineMsg)
	} else {
		return 0
	}
}

type ClientOfflineHandle struct {
	ClientOfflineNum int `json:"clientofflinenum"`
}

func (list *ClientOfflineHandle)ServeHTTP(w http.ResponseWriter, r *http.Request) {

	arg1 := r.URL.Query().Get("clientId")
	list.ClientOfflineNum = GetClientOffLineMsgList(arg1)
	jsonResp, err := json.Marshal(list)
	if err != nil {
		fmt.Println(err)
		return
	}
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Content-Type", "application/json")
	w.Write(jsonResp)
}
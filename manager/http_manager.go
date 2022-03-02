package manager

import (
	"fmt"
	"github.com/mqtt_server/MQTT_Server_Go/config"
	"net/http"
)

var mqttServerMux http.ServeMux
var onLineNumHandle OnLineNumHandle
var clientListHandle ClientInfoListHandle
var clientSubHandle ClientSubInfoHandle
var clientOffMsgHandle ClientOfflineHandle

func Http_Manager_Server() {

	mqttServerMux.Handle("/getOnLineNum", &onLineNumHandle)
	mqttServerMux.Handle("/getClientList", &clientListHandle)
	mqttServerMux.Handle("/getClientSub", &clientSubHandle)
	mqttServerMux.Handle("/getOfflienMsg", &clientOffMsgHandle)
	port, err := config.ReadConfigValueInt("http", "port")
	if err != nil {
		port = 9000
	}
	address := fmt.Sprintf("127.0.0.1:%d",port)
	server := &http.Server{Addr: address, Handler: &mqttServerMux}
	server.ListenAndServe()

}

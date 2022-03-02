package manager

import (
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

	server := &http.Server{Addr: "127.0.0.1:9000", Handler: &mqttServerMux}
	server.ListenAndServe()

}
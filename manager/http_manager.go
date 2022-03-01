package manager

import (
	"net/http"
)


var mqttServerMux http.ServeMux
var onLineNumHandle OnLineNumHandle
var clientListHandle ClientInfoListHandle

func Http_Manager_Server() {

	mqttServerMux.Handle("/getOnLineNum", &onLineNumHandle)
	mqttServerMux.Handle("/getClientList", &clientListHandle)

	server := &http.Server{Addr: "127.0.0.1:9000", Handler: &mqttServerMux}
	server.ListenAndServe()

}
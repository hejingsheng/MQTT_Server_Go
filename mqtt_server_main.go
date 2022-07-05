package main

import (
	"flag"
	"fmt"
	"mqtt_server/MQTT_Server_Go/cluster"
	"mqtt_server/MQTT_Server_Go/config"
	"mqtt_server/MQTT_Server_Go/log"
	"mqtt_server/MQTT_Server_Go/manager"
	"mqtt_server/MQTT_Server_Go/process"
	"net"
)

var SOFTWARE_VERSION string = "V0.0.1"
var SOFTWARE_AUTHOR  string = "Jack He"

var c = flag.String("c", "", "Config File Path")

func parseArgs() string {
	flag.Parse()
	return *c
}

func main() {
	mainThreadId := "main"
	log.LogInit(log.LOG_DEBUG)
	log.LogPrint(log.LOG_INFO, "[%s] MQTT Server Version:%s, Author:%s", mainThreadId, SOFTWARE_VERSION, SOFTWARE_AUTHOR)
	log.LogPrint(log.LOG_INFO, "[%s] MQTT Server Running...", mainThreadId)

	configfile := parseArgs()
	config.ConfigFileInit(configfile)

	port, err := config.ReadConfigValueInt("mqtt", "port")
	if err != nil {
		port = 1883
	}
	address := fmt.Sprintf("0.0.0.0:%d",port)
	listen, err := net.Listen("tcp", address)
	if err != nil {
		log.LogPrint(log.LOG_ERROR, "[%s] listen error", mainThreadId)
		return
	}
	defer listen.Close()

	dispatchChannal := make(chan process.DispatchRoutinMsg)
	go process.ClientsMsgDispatch(dispatchChannal)
	go manager.Http_Manager_Server()
	cluster.StartClusterServerNode(dispatchChannal)

	for {
		log.LogPrint(log.LOG_INFO, "[%s] wait a client connect", mainThreadId)
		conn, err := listen.Accept()
		if err != nil {
			log.LogPrint(log.LOG_ERROR, "[%s] client connect error", mainThreadId)
		} else {
			log.LogPrint(log.LOG_INFO, "[%s] a client connect success %v", mainThreadId, conn.RemoteAddr().String())
			go process.ClientProcess(conn, dispatchChannal)
		}

	}
}

package main

import (
	"github.com/mqtt_server/MQTT_Server_Go/log"
	"github.com/mqtt_server/MQTT_Server_Go/process"
	"net"
)

func main() {
	mainThreadId := "main"
	log.LogInit(log.LOG_INFO)
	log.LogPrint(log.LOG_INFO, mainThreadId, "MQTT Server Running...")

	listen, err := net.Listen("tcp", "0.0.0.0:1883")
	if err != nil {
		log.LogPrint(log.LOG_ERROR, mainThreadId, "listen error")
		return
	}
	defer listen.Close()

	for {
		log.LogPrint(log.LOG_INFO, mainThreadId, "wait a client connect")
		conn, err := listen.Accept()
		if err != nil {
			log.LogPrint(log.LOG_ERROR, mainThreadId, "client connect error")
		} else {
			log.LogPrint(log.LOG_INFO, mainThreadId, "a client connect success %v", conn.RemoteAddr().String())
			go process.ClientProcess(conn)
		}

	}
}

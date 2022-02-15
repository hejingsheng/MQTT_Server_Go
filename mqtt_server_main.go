package main

import (
	"github.com/mqtt_server/MQTT_Server_Go/log"
	"github.com/mqtt_server/MQTT_Server_Go/process"
	"net"
)

func main() {
	log.LogInit(log.LOG_DEBUG)
	log.LogPrint(log.LOG_DEBUG, "MQTT Server Running...")

	listen, err := net.Listen("tcp", "0.0.0.0:1883");
	if err != nil {
		log.LogPrint(log.LOG_ERROR, "listen error")
		return;
	}
	defer listen.Close()

	for {
		log.LogPrint(log.LOG_INFO, "wait a client connect")
		conn, err := listen.Accept()
		if err != nil {
			log.LogPrint(log.LOG_ERROR, "client connect error")
		} else {
			log.LogPrint(log.LOG_INFO, "a client connect success %v", conn.RemoteAddr().String())
			go process.ClientProcess(conn)
		}

	}
}
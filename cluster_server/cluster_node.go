package cluster_server

import (
	"mqtt_server/MQTT_Server_Go/config"
	"mqtt_server/MQTT_Server_Go/log"
	"net"
	"os"
	"strconv"
	"strings"
)

func startClusterServer(address string) {
	udpAddr, err := net.ResolveUDPAddr("udp", address)
	conn, err := net.ListenUDP("udp", udpAddr)
	defer conn.Close()

	if err != nil {
		log.LogPrint(log.LOG_PANIC, "Slave UDP Server Start Error %v", err)
		os.Exit(1)
	}

	for {
		data := make([]byte, 4096)
		n, _, err := conn.ReadFromUDP(data)
		if err != nil {
			log.LogPrint(log.LOG_PANIC, "Failed To Read UDP Msg, Error: %v", err)
		}
		log.LogPrint(log.LOG_DEBUG, "Read UDP Msg, Len: %v", n)
	}
}

func StartClusterServerNode() int {
	cluster_nodes, err := config.ReadConfigValueString("cluster", "coservers")
	if err != nil {
		return 0
	}
	serverNodes := strings.Split(cluster_nodes, "/")
	for _, node := range serverNodes {
		log.LogPrint(log.LOG_DEBUG, "Get one server node [%s]", node)
	}
	ClusterPort,_ := config.ReadConfigValueInt("cluster", "port")
	address := "0.0.0.0:"+strconv.Itoa(ClusterPort)
	go startClusterServer(address)
	return 1
}
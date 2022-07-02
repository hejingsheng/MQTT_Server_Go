package cluster

import (
	"mqtt_server/MQTT_Server_Go/config"
	"mqtt_server/MQTT_Server_Go/log"
	"net"
	"strconv"
	"strings"
	"time"
)

const (
	TIME_1SECOND = 100
	TIME_1MIN = 100 *60
)

func startClusterServerRecv(conn *net.UDPConn) {

	defer conn.Close()

	for {
		data := make([]byte, 4096)
		n, _, err := conn.ReadFromUDP(data)
		if err != nil {
			log.LogPrint(log.LOG_PANIC, "Failed To Read UDP Msg, Error: %v", err)
			break
		}
		log.LogPrint(log.LOG_DEBUG, "Read UDP Msg, Len: %v", n)
	}
}

func startClusterServerCycly(conn *net.UDPConn) {
	timeEscape100MS := 0
	ticker := time.NewTicker(time.Millisecond * 100)

	for {
		select {
		case <-ticker.C:
			timeEscape100MS++
			if timeEscape100MS == TIME_1SECOND {

			}
		}
	}
	ticker.Stop()
}

func createClusterSocket(address string) (*net.UDPConn, error) {
	udpAddr, err := net.ResolveUDPAddr("udp", address)
	if err != nil {
		return nil, err
	}
	conn, err := net.ListenUDP("udp", udpAddr)
	if err != nil {
		return nil, err
	}
	return conn, nil
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
	conn, err := createClusterSocket(address)
	if err != nil {
		return 0
	}
	go startClusterServerRecv(conn)
	go startClusterServerCycly(conn)
	return 1
}
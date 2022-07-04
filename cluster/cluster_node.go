package cluster

import (
	"net"
	"strconv"
	"strings"
)

type ClusterNodeInfo struct {
	Ip               string
	Port             uint16
	KeepAliveTimeOut int8
	HeartTime        uint8
}

func CreateClusterNode(address string) *ClusterNodeInfo {
	node := new(ClusterNodeInfo)
	index := strings.Index(address, ":")
	node.Ip = address[0:index]
	port, _ := strconv.Atoi(address[index+1 : len(address)])
	node.Port = uint16(port)
	node.KeepAliveTimeOut = 60
	node.HeartTime = 0
	return node
}

func (node *ClusterNodeInfo) SendMsgToNode(conn *net.UDPConn, data []byte) (int, error) {
	var addr net.UDPAddr
	addr.Port = int(node.Port)
	addr.IP = net.ParseIP(node.Ip).To4()
	ret, err := conn.WriteTo(data, &addr)
	if err != nil {
		return -1, err
	}
	return ret, nil
}

func (node *ClusterNodeInfo) UpdateKeepAliveTimeout() {
	node.KeepAliveTimeOut = 60
}

func (node *ClusterNodeInfo) KeepAliveTimeEscape1Sec() (int8, bool) {
	node.KeepAliveTimeOut--
	if node.KeepAliveTimeOut <= 0 {
		node.KeepAliveTimeOut = 0
		return 0, false
	}
	return node.KeepAliveTimeOut, true
}

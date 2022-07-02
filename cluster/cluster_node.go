package cluster

import (
	"net"
	"strconv"
	"strings"
)

type ClusterNodeInfo struct {
	Ip string
	Port uint16
}

func CreateClusterNode(address string) *ClusterNodeInfo {
	node := new(ClusterNodeInfo)
	index := strings.Index(address, ":")
	node.Ip = address[0:index]
	port, _ := strconv.Atoi(address[index+1 : len(address)])
	node.Port = uint16(port)
	return node
}

func (node *ClusterNodeInfo)SendMsgToNode(conn *net.UDPConn, data []byte) (int, error) {
	var addr net.UDPAddr
	addr.Port = int(node.Port)
	addr.IP = net.ParseIP(node.Ip).To4()
	ret, err := conn.WriteTo(data, &addr)
	if err != nil {
		return -1, err
	}
	return ret, nil
}

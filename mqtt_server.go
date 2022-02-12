package main

import (
	"fmt"
	"net"
	"sync"
)

var (
	gloablClientsMap map[net.Conn]int
	globalClientsMapLock sync.Mutex
)

func tcpClientProcess(localconn net.Conn) {
	fmt.Println("start a tcp client")

	defer func() {
		fmt.Println("one client close exit process")
		globalClientsMapLock.Lock()
		delete(gloablClientsMap, localconn)
		globalClientsMapLock.Unlock()
		localconn.Close()
	}();

	for {
		var read_buf []byte = make([]byte, 1024)
		num, err := localconn.Read(read_buf)
		if err != nil {
			fmt.Println("read error")
			break
		} else {
			fmt.Printf("read %d data %s\n", num, read_buf)
			var write_buf []byte = make([]byte, num)
			for i := 0; i < num; i++ {
				write_buf[i] = read_buf[i];
			}
			for conn, index := range gloablClientsMap {
				//fmt.Println(country, "首都是", countryCapitalMap [country])
				if conn == localconn{
					continue
				}
				fmt.Println("this is ", index, " conn")
				conn.Write(write_buf);
			}
		}
	}
}

func main() {
	fmt.Println("MQTT Server Running...")
	var index int = 1
	gloablClientsMap = make(map[net.Conn]int, 1)
	listen, err := net.Listen("tcp", "0.0.0.0:1883");
	if err != nil {
		fmt.Println("listen error");
		return;
	}
	defer listen.Close()

	for {
		fmt.Println("wait a client connect")
		conn, err := listen.Accept()
		if err != nil {
			fmt.Println("client connect error")
		} else {
			fmt.Printf("a client connect success %v\n", conn.RemoteAddr().String())
			globalClientsMapLock.Lock()
			gloablClientsMap[conn] = index
			globalClientsMapLock.Unlock()
			index++;
			go tcpClientProcess(conn)
		}

	}
}
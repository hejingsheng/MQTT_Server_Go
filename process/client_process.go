package process

import (
	"fmt"
	"net"
	"sync"
)

var (
	gloablClientsMap map[net.Conn]int
	globalClientsMapLock sync.Mutex
	index int = 1
)

func init() {
	gloablClientsMap = make(map[net.Conn]int, 1)
}

func ClientProcess(localconn net.Conn) {
	fmt.Println("start a tcp client")

	globalClientsMapLock.Lock()
	gloablClientsMap[localconn] = index
	globalClientsMapLock.Unlock()

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
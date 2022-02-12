package main

import (
	"fmt"
	"net"
)

func main() {
	fmt.Println("MQTT Server Running...")

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
		}

	}
}
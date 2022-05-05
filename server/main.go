package main

import (
	"log"
	"net"

	"github.com/Mrtoy/dtcg-server/service"
)

func main() {
	s := service.NewServer()
	listener, err := net.Listen("tcp", "0.0.0.0:2333")
	if err != nil {
		log.Fatalln("Listen tcp server failed,err:", err)
	}

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Println("Listen.Accept failed,err:", err)
			continue
		}
		go s.HandleConnect(conn)
	}
}

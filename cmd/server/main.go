package main

import (
	"github.com/Mrtoy/dtcg-server/controller"
	"github.com/Mrtoy/dtcg-server/service"
)

func main() {
	s := service.NewServer()
	controller.Route(s)
	s.Listen(":2333")
}

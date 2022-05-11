package main

import (
	"github.com/Mrtoy/dtcg-server/app"
	"github.com/Mrtoy/dtcg-server/controller"
	"github.com/Mrtoy/dtcg-server/service"
)

func main() {
	s := service.NewServer()
	g := controller.NewGameController()
	r := controller.NewRoomController()
	s.On("connect", func(pack *service.Package, sess *service.Session) {
		app.SetPlayer(sess)
		sess.Send("connect", map[string]any{
			"ID": sess.ID,
		})
	})
	s.On("disconnect", func(pack *service.Package, sess *service.Session) {
		r.Leave(pack, sess)
		g.Leave(pack, sess)
	})
	s.On("player:update-info", controller.UpdatePlayerInfo)
	s.On("room:join", r.Join)
	s.On("room:leave", r.Leave)
	s.On("room:ready", r.Ready)
	s.On("game:born", g.Born)
	s.Listen(":2333")
}

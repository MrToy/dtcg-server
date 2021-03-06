package main

import (
	"github.com/Mrtoy/dtcg-server/app"
	"github.com/Mrtoy/dtcg-server/controller"
	"github.com/Mrtoy/dtcg-server/service"
)

func main() {
	s := service.NewServer()
	r := controller.NewRoomController()
	s.On("connect", func(pack *service.Package, sess *service.Session) {
		app.SetPlayer(sess)
		sess.Send("connect", map[string]any{
			"ID": sess.ID,
		})
	})
	s.On("disconnect", func(pack *service.Package, sess *service.Session) {
		r.Leave(pack, sess)
		controller.LeaveGame(pack, sess)
		r.LeaveSolo(pack, sess)
	})
	s.On("player:update-info", controller.UpdatePlayerInfo)

	s.On("room:join", r.Join)
	s.On("room:leave", r.Leave)
	s.On("room:join-solo", r.JoinSolo)
	s.On("room:leave-solo", r.LeaveSolo)
	s.On("room:ready", r.Ready)

	s.On("game:*", controller.OnGameMessage)
	s.Listen(":2333")
}

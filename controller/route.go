package controller

import (
	"github.com/Mrtoy/dtcg-server/app"
	"github.com/Mrtoy/dtcg-server/service"
)

func Route(s *service.Server) {
	g := NewGameController()
	r := NewRoomController()
	s.On("connect", func(pack *service.Package, sess *service.Session) {
		app.SetPlayer(sess)
		sess.Send("connect", map[string]any{
			"ID": sess.ID,
		})
	})
	s.On("disconnect", func(pack *service.Package, sess *service.Session) {
		r.Leave(pack, sess)
		g.ExitGame(pack, sess)
	})
	s.On("player:update-info", UpdatePlayerInfo)
	s.On("room:join", r.Join)
	s.On("room:leave", r.Leave)
	s.On("room:ready", r.Ready)
	s.On("game:born", g.Born)
}

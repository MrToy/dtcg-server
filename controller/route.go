package controller

import (
	"github.com/Mrtoy/dtcg-server/app"
	"github.com/Mrtoy/dtcg-server/service"
)

func Route(s *service.Server) {
	gameManager := NewGameManager()
	roomManger := NewRoomManager()
	roomManger.GameManager = gameManager
	s.On("connect", func(pack *service.Package, sess *service.Session) {
		app.SetPlayer(sess)
		sess.Send("connect", map[string]any{
			"ID": sess.ID,
		})
	})
	s.On("disconnect", func(pack *service.Package, sess *service.Session) {
		roomManger.Leave(pack, sess)
		gameManager.ExitGame(pack, sess)
	})
	s.On("player:update-info", UpdatePlayerInfo)
	s.On("room:join", roomManger.Join)
	s.On("room:leave", roomManger.Leave)
	s.On("room:ready", roomManger.Ready)
	s.On("game:born", gameManager.Born)
}

package main

import (
	"github.com/Mrtoy/dtcg-server/server"
	"github.com/Mrtoy/dtcg-server/service"
)

func main() {
	s := server.NewServer()
	gameManager := service.NewGameManager()
	roomManger := service.NewRoomManager()
	roomManger.GameManager = gameManager
	s.On("connect", func(pack *server.Package, sess *server.Session) {
		service.SetPlayer(sess)
		sess.Send("connect", map[string]any{
			"ID": sess.ID,
		})
	})
	s.On("disconnect", func(pack *server.Package, sess *server.Session) {
		roomManger.Leave(pack, sess)
		gameManager.ExitGame(pack, sess)
	})

	s.On("player:update-info", service.UpdatePlayerInfo)

	s.On("room:join", roomManger.Join)
	s.On("room:leave", roomManger.Leave)
	s.On("room:ready", roomManger.Ready)

	s.On("game:born", gameManager.Born)

	s.Listen(":2333")
}

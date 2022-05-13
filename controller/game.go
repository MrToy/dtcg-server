package controller

import (
	"github.com/Mrtoy/dtcg-server/app"
	"github.com/Mrtoy/dtcg-server/service"
)

func StartGame(players []*app.Player) {
	g := app.NewGame(players)
	for _, p := range players {
		p.Session.Data["game"] = g
	}
	go g.Start()
}

func LeaveGame(pack *service.Package, sess *service.Session) {
	player := sess.Data["player"].(*app.Player)
	g, ok := sess.Data["game"].(*app.Game)
	if !ok {
		return
	}
	g.WinPlayer = player.Opponent
	g.EndChan <- true
}

func OnGameMessage(pack *service.Package, sess *service.Session) {
	g, ok := sess.Data["game"].(*app.Game)
	if !ok {
		return
	}
	g.MessageChan <- pack
}

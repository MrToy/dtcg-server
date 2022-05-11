package controller

import (
	"github.com/Mrtoy/dtcg-server/app"
	"github.com/Mrtoy/dtcg-server/service"
)

type GameController struct {
}

func NewGameController() *GameController {
	return &GameController{}
}

func StartGame(players []*app.Player) {
	g := app.NewGame(players)
	for _, p := range players {
		p.Session.Data["game"] = g
	}
	go g.Start()
}

func (m *GameController) Leave(pack *service.Package, sess *service.Session) {
	player := sess.Data["player"].(*app.Player)
	g := sess.Data["game"].(*app.Game)
	if g == nil {
		return
	}
	g.WinPlayer = player.Opponent
	g.End()
}

func (m *GameController) Born(pack *service.Package, sess *service.Session) {
	g := sess.Data["game"].(*app.Game)
	if g == nil {
		return
	}
	g.Born()
}

func (m *GameController) SummonBorn(pack *service.Package, sess *service.Session) {
	g := sess.Data["game"].(*app.Game)
	if g == nil {
		return
	}
	g.SummonBorn()
}

func (m *GameController) Summon(pack *service.Package, sess *service.Session) {
	g := sess.Data["game"].(*app.Game)
	if g == nil {
		return
	}
	var req map[string]any
	pack.Unmarshal(&req)
	id := req["id"].(int)
	g.Summon(id)
}

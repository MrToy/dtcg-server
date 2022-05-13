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

func WithGameInfo(fn func(*service.Package, *service.Session, *app.Game, *app.PlayerArea)) func(*service.Package, *service.Session) {
	return func(pack *service.Package, sess *service.Session) {
		player := sess.Data["player"].(*app.Player)
		g, ok := sess.Data["game"].(*app.Game)
		if !ok {
			return
		}
		if g.OperationPlayer != player {
			return
		}
		area := g.PlayerAreas[player.Session.ID]
		if area == nil {
			return
		}
		fn(pack, sess, g, area)
	}
}

func (m *GameController) Leave(pack *service.Package, sess *service.Session, g *app.Game, area *app.PlayerArea) {
	player := sess.Data["player"].(*app.Player)
	g.WinPlayer = player.Opponent
	g.EndGame()
}

func (m *GameController) Born(pack *service.Package, sess *service.Session, g *app.Game, area *app.PlayerArea) {
	if len(area.Born) == 0 {
		g.Born()
	} else {
		g.BornSummon()
	}
}

func (m *GameController) PlayCard(pack *service.Package, sess *service.Session, g *app.Game, area *app.PlayerArea) {
	var req struct {
		ID int `json:"id"`
	}
	pack.Unmarshal(&req)
	g.Summon(req.ID)
}

func (m *GameController) Attack(pack *service.Package, sess *service.Session, g *app.Game, area *app.PlayerArea) {
	var req struct {
		ID int `json:"id"`
	}
	pack.Unmarshal(&req)
	g.Summon(req.ID)
}

func (m *GameController) NextTurn(pack *service.Package, sess *service.Session, g *app.Game, area *app.PlayerArea) {
	g.TurnChan <- true
}

package controller

import (
	"github.com/Mrtoy/dtcg-server/app"
	"github.com/Mrtoy/dtcg-server/service"
)

type GameManager struct {
	PlayerInGame map[int]*app.Game
}

func NewGameManager() *GameManager {
	return &GameManager{
		PlayerInGame: make(map[int]*app.Game),
	}
}

func (m *GameManager) StartGame(room *app.Room) *app.Game {
	g := &app.Game{
		Room:        room,
		MemoryBank:  0,
		TurnChan:    make(chan bool),
		EndChan:     make(chan bool),
		PlayerAreas: make(map[int]*app.PlayerArea),
	}
	for _, player := range room.Players {
		area := &app.PlayerArea{}
		g.PlayerAreas[player.Session.ID] = area
		m.PlayerInGame[player.Session.ID] = g
	}
	g.Players[0].Opponent = g.Players[1]
	g.Players[1].Opponent = g.Players[0]
	go g.Start()
	return g
}

func (m *GameManager) ExitGame(pack *service.Package, sess *service.Session) {
	player := sess.Data["player"].(*app.Player)
	g := m.PlayerInGame[player.Session.ID]
	if g != nil {
		g.WinPlayer = player.Opponent
		g.End()
	}
	delete(m.PlayerInGame, player.Session.ID)
}

func (m *GameManager) Born(pack *service.Package, sess *service.Session) {
	player := sess.Data["player"].(*app.Player)
	g := m.PlayerInGame[player.Session.ID]
	if g == nil {
		return
	}
	g.Born()
}

func (m *GameManager) SummonBorn(pack *service.Package, sess *service.Session) {
	player := sess.Data["player"].(*app.Player)
	g := m.PlayerInGame[player.Session.ID]
	if g == nil {
		return
	}
	g.SummonBorn()
}

func (m *GameManager) Summon(pack *service.Package, sess *service.Session) {
	player := sess.Data["player"].(*app.Player)
	g := m.PlayerInGame[player.Session.ID]
	if g == nil {
		return
	}
	var req map[string]any
	pack.Unmarshal(&req)
	id := req["id"].(int)
	g.Summon(id)
}

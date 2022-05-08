package service

import (
	"errors"
	"time"

	"github.com/Mrtoy/dtcg-server/server"
)

var ErrInvalidOperation = errors.New("invalid operation")

type PlayerArea struct {
	Egg     []Card         //数码蛋
	Born    CardMonster    //育成区
	Hand    []Card         //手牌
	Field   []*CardMonster //场上
	Deck    []Card         //卡组
	Discard []Card         //弃牌堆
	Defense []Card         //安防区
}

type Game struct {
	*Room
	PlayerAreas     []PlayerArea
	MemoryBank      int //内存条
	CurrentPlayer   *Player
	OperationPlayer *Player
	TurnChan        chan bool `json:"-"`
	EndChan         chan bool `json:"-"`
	WinPlayer       *Player
}

func (g *Game) Start() {
	g.Players[0].Opponent = g.Players[1]
	g.Players[0].PlayerArea = &g.PlayerAreas[0]
	g.Players[1].Opponent = g.Players[0]
	g.Players[1].PlayerArea = &g.PlayerAreas[1]

	g.BroadCast("game:start", nil)

	g.SetupPlayer(&g.PlayerAreas[0], g.Players[0])
	g.SetupPlayer(&g.PlayerAreas[1], g.Players[1])

	g.CurrentPlayer = g.Players[0]
	g.OperationPlayer = g.CurrentPlayer
	g.OnCurrentChange()

	turnList := []string{"active", "draw", "born", "main", "end"}
	turnIndex := 0

	for {
		turnIndex++
		if turnIndex > len(turnList)-1 {
			turnIndex = 0
		}
		currentTurn := turnList[turnIndex]
		go g.OnTurn(currentTurn)
		select {
		case <-g.EndChan:
			return
		case <-g.TurnChan:
		case <-time.After(30 * time.Second):
		}
	}
}

func (g *Game) End() {
	g.BroadCast("game:end", map[string]any{
		"winner": g.WinPlayer.Session.ID,
	})
	g.EndChan <- true
}

func (g *Game) SetupPlayer(area *PlayerArea, player *Player) {
	ShuffleCards(player.EggDeck)
	ShuffleCards(player.Deck)
	area.Egg = player.EggDeck
	area.Deck = player.Deck
	area.Deck, area.Defense, _ = MoveCards(area.Deck, area.Defense, 5)
	area.Deck, area.Hand, _ = MoveCards(area.Deck, area.Hand, 5)
}

func (g *Game) OnCurrentChange() {
	g.BroadCast("game:current-change", map[string]any{
		"current":   g.CurrentPlayer.Session.ID,
		"operation": g.OperationPlayer.Session.ID,
	})
}

func (g *Game) OnTurn(currentTurn string) {
	g.BroadCast("game:turn-change", currentTurn)
	switch currentTurn {
	case "active":
		g.OnTurnActive()
	case "draw":
		g.OnTurnDraw()
	case "born":
		g.OnTurnBorn()
	case "main":
		g.OnTurnMain()
	case "end":
		g.OnTurnEnd()
	}
}

func (g *Game) OnTurnActive() {

	g.TurnChan <- true
	// g.BroadCastEvent(&Event{Type: "game:turn-draw", Target: g.CurrentPlayer})
}

func (g *Game) OnTurnDraw() {

	g.TurnChan <- true
	// g.BroadCastEvent(&Event{Type: "game:turn-draw", Target: g.CurrentPlayer})
}

func (g *Game) OnTurnBorn() {
	g.CurrentPlayer.Session.Send("confirm", map[string]any{
		"message":  "是否需要育成",
		"callback": "game:born",
	})
}

func (g *Game) OnTurnMain() {

	// g.BroadCastEvent(&Event{Type: "game:turn-main", Target: g.CurrentPlayer})
}

func (g *Game) OnTurnEnd() {
	// g.BroadCastEvent(&Event{Type: "game:turn-end", Target: g.CurrentPlayer})
	g.CurrentPlayer = g.CurrentPlayer.Opponent
	g.OperationPlayer = g.CurrentPlayer
	g.OnCurrentChange()
	g.TurnChan <- true
	// g.BroadCastEvent(&Event{Type: "game:operation-change", Target: g.OperationPlayer})
}

type GameManager struct {
	PlayerInGame map[int]*Game
}

func NewGameManager() *GameManager {
	return &GameManager{
		PlayerInGame: make(map[int]*Game),
	}
}

func (m *GameManager) StartGame(room *Room) *Game {
	g := &Game{
		Room:        room,
		PlayerAreas: make([]PlayerArea, 2),
		MemoryBank:  0,
		TurnChan:    make(chan bool),
		EndChan:     make(chan bool),
	}
	for _, player := range room.Players {
		m.PlayerInGame[player.Session.ID] = g
	}

	go g.Start()
	return g
}

func (m *GameManager) ExitGame(pack *server.Package, sess *server.Session) {
	player := sess.Data["player"].(*Player)
	g := m.PlayerInGame[player.Session.ID]
	if g != nil {
		g.WinPlayer = player.Opponent
		g.End()
	}
	delete(m.PlayerInGame, player.Session.ID)
}

func (m *GameManager) Born(pack *server.Package, sess *server.Session) {
	player := sess.Data["player"].(*Player)
	g := m.PlayerInGame[player.Session.ID]
	if g == nil {
		return
	}
	g.TurnChan <- true
}

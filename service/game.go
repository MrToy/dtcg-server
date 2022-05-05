package service

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"time"
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
	Players         []*Player
	PlayerAreas     []PlayerArea
	MemoryBank      int //内存条
	CurrentPlayer   *Player
	OperationPlayer *Player
	TurnChan        chan bool `json:"-"`
	EndChan         chan bool `json:"-"`
	WinPlayer       *Player
}

func NewGame() *Game {
	g := &Game{
		PlayerAreas: make([]PlayerArea, 2),
		MemoryBank:  0,
		TurnChan:    make(chan bool),
		EndChan:     make(chan bool),
	}
	return g
}

func (g *Game) Start() {
	g.Players[0].Opponent = g.Players[1]
	g.Players[0].PlayerArea = &g.PlayerAreas[0]
	g.Players[1].Opponent = g.Players[0]
	g.Players[1].PlayerArea = &g.PlayerAreas[1]

	g.BroadCastEvent(&Event{Type: "game:start"})

	g.SetupPlayer(&g.PlayerAreas[0], g.Players[0])
	g.SetupPlayer(&g.PlayerAreas[1], g.Players[1])

	//这里需要猜拳
	g.CurrentPlayer = g.Players[0]
	g.OperationPlayer = g.CurrentPlayer
	for {
		g.OnTurnActive()
		g.OnTurnDraw()
		g.OnTurnBorn()
		select {
		case <-g.EndChan:
			return
		case <-g.TurnChan:
		case <-time.After(233 * time.Second):
		}
		g.OnTurnMain()
		select {
		case <-g.EndChan:
			return
		case <-g.TurnChan:
		case <-time.After(233 * time.Second):
		}
		g.OnTurnEnd()
	}
}

func (g *Game) End() {
	log.Println("OnGameEnd")
	g.BroadCastEvent(&Event{Type: "game:win", Target: g.WinPlayer})
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

func (g *Game) HandleEvent(e *Event) error {
	if e.Type != "game:surrender" && e.Target != g.OperationPlayer {
		return ErrInvalidOperation
	}
	switch e.Type {
	case "game:surrender":
		g.WinPlayer = e.Target.Opponent
		g.End()
	case "game:next-turn":
		g.TurnChan <- true
	}
	return nil
}

func (g *Game) OnTurnActive() {
	g.BroadCastEvent(&Event{Type: "game:turn-active", Target: g.CurrentPlayer})
}

func (g *Game) OnTurnDraw() {
	g.BroadCastEvent(&Event{Type: "game:turn-draw", Target: g.CurrentPlayer})
}

func (g *Game) OnTurnBorn() {
	g.BroadCastEvent(&Event{Type: "game:turn-born", Target: g.CurrentPlayer})
}

func (g *Game) OnTurnMain() {
	g.BroadCastEvent(&Event{Type: "game:turn-main", Target: g.CurrentPlayer})
}

func (g *Game) OnTurnEnd() {
	g.BroadCastEvent(&Event{Type: "game:turn-end", Target: g.CurrentPlayer})
	g.CurrentPlayer = g.CurrentPlayer.Opponent
	g.OperationPlayer = g.CurrentPlayer
	g.BroadCastEvent(&Event{Type: "game:operation-change", Target: g.OperationPlayer})
}

func (g *Game) BroadCastState() {
	fmt.Println(json.Marshal(g))
}

func (g *Game) BroadCastEvent(e *Event) {
	for _, p := range g.Players {
		p.WriteChan <- e
	}
}

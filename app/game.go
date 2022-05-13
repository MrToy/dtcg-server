package app

import (
	"encoding/json"
	"errors"
	"log"
	"time"

	"github.com/sergi/go-diff/diffmatchpatch"
)

var ErrInvalidOperation = errors.New("invalid operation")

type PlayerArea struct {
	Egg     []*Card        //数码蛋
	Born    []*MonsterCard //育成区
	Hand    []*Card        //手牌
	Field   []*MonsterCard //场上
	Deck    []*Card        //卡组
	Discard []*Card        //弃牌堆
	Defense []*Card        //安防区
}

func NewPlayerArea() *PlayerArea {
	return &PlayerArea{
		Egg:     []*Card{},
		Born:    []*MonsterCard{},
		Hand:    []*Card{},
		Field:   []*MonsterCard{},
		Deck:    []*Card{},
		Discard: []*Card{},
		Defense: []*Card{},
	}
}

type Game struct {
	Players           []*Player
	PlayerAreas       map[int]*PlayerArea
	MemoryBank        int //内存条
	CurrentPlayer     *Player
	OperationPlayer   *Player
	TurnChan          chan bool `json:"-"`
	EndChan           chan bool `json:"-"`
	WinPlayer         *Player
	PreStateStr       string `json:"-"`
	CurrentTurn       string
	InstanceIDCounter int `json:"-"`
	GameEnd           bool
}

func NewGame(players []*Player) *Game {
	gamePlayers := make([]*Player, len(players))
	copy(gamePlayers, players)
	g := &Game{
		Players:     gamePlayers,
		MemoryBank:  0,
		TurnChan:    make(chan bool),
		EndChan:     make(chan bool),
		PlayerAreas: make(map[int]*PlayerArea),
	}
	for _, player := range g.Players {
		area := NewPlayerArea()
		g.PlayerAreas[player.Session.ID] = area
	}
	g.Players[0].Opponent = g.Players[1]
	g.Players[1].Opponent = g.Players[0]
	return g
}

func (g *Game) Broadcast(tp string, data any) {
	for _, p := range g.Players {
		err := p.Session.Send(tp, data)
		if err != nil {
			log.Println("send error:", err)
		}
	}
}

func (g *Game) BroadcastGameInfo() {
	dmp := diffmatchpatch.New()
	bytes, _ := json.Marshal(g)
	patchs := dmp.PatchMake(g.PreStateStr, string(bytes))
	g.PreStateStr = string(bytes)
	g.Broadcast("game:info-diff", dmp.PatchToText(patchs))
	// log.Println(dmp.PatchApply(patchs, g.PreStateStr))
}

func (g *Game) BroadcastDeckDetails() {
	details := map[string]*CardDetail{}
	for _, p := range g.Players {
		for _, s := range p.OriginDeck {
			details[s] = GetDetail(s)
		}
	}
	g.Broadcast("game:deck-details", details)
}

func (g *Game) Start() {
	g.Broadcast("game:start", nil)
	g.BroadcastDeckDetails()

	g.CurrentPlayer = g.Players[0]

	g.SetupDeck(g.Players[0])
	g.SetupDeck(g.Players[1])
	g.BroadcastGameInfo()

	g.SetupDraw(g.Players[0])
	g.SetupDraw(g.Players[1])
	g.BroadcastGameInfo()

	g.CurrentTurn = "born"

	for {
		if g.GameEnd {
			g.onGameEnd()
			return
		}
		g.Update()
		if g.OperationPlayer != nil {
			log.Printf("player %d -- user action", g.OperationPlayer.Session.ID)
			select {
			case <-g.TurnChan:
			case <-time.After(233 * time.Second):
			}
			g.OperationPlayer = nil
		}
	}
}

func (g *Game) Update() {
	log.Printf("player %d -- turn %s", g.CurrentPlayer.Session.ID, g.CurrentTurn)
	switch g.CurrentTurn {
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
	g.BroadcastGameInfo()

	index := ListFindIndex(turnList, func(s string) bool {
		return s == g.CurrentTurn
	})
	if index == len(turnList)-1 {
		g.CurrentTurn = turnList[0]
	} else {
		g.CurrentTurn = turnList[index+1]
	}
}

func (g *Game) onGameEnd() {
	g.BroadcastGameInfo()
	g.Broadcast("game:end", nil)
	for _, p := range g.Players {
		delete(p.Session.Data, "game")
	}
}

func (g *Game) SetupDeck(p *Player) {
	deck := []*Card{}
	for _, s := range p.OriginDeck {
		card := NewCard()
		card.Serial = s
		card.ID = g.InstanceIDCounter
		g.InstanceIDCounter++
		deck = append(deck, card)
	}
	ShuffleCards(deck)
	g.PlayerAreas[p.Session.ID].Deck = deck
}

func (g *Game) SetupDraw(player *Player) {
	area := g.PlayerAreas[player.Session.ID]
	var picked []*Card
	picked, area.Deck = ListPickSome(area.Deck, func(a *Card) bool {
		d := GetDetail(a.Serial)
		return d.Level == "2"
	})
	area.Egg = ListAppend(area.Egg, picked)
	area.Deck, area.Defense, _ = ListMove(area.Deck, area.Defense, 5)
	area.Deck, area.Hand, _ = ListMove(area.Deck, area.Hand, 5)
}

var turnList = []string{"active", "draw", "born", "main", "end"}

func (g *Game) OnTurnActive() {

}

func (g *Game) OnTurnDraw() {
	area := g.PlayerAreas[g.CurrentPlayer.Session.ID]
	var err error
	area.Deck, area.Hand, err = ListMove(area.Deck, area.Hand, 1)
	if err != nil {
		g.WinPlayer = g.CurrentPlayer.Opponent
		g.GameEnd = true
		return
	}
}

func (g *Game) OnTurnBorn() {
	message := "是否需要育成?"
	area := g.PlayerAreas[g.CurrentPlayer.Session.ID]
	if len(area.Born) > 0 {
		message = "是否登场育成区怪兽?"
	}
	g.CurrentPlayer.Session.Send("confirm", map[string]any{
		"message":  message,
		"callback": "game:born",
	})
	g.OperationPlayer = g.CurrentPlayer
}

func (g *Game) OnTurnMain() {
	g.OperationPlayer = g.CurrentPlayer
}

func (g *Game) OnTurnEnd() {
	g.CurrentPlayer = g.CurrentPlayer.Opponent
	g.BroadcastGameInfo()
}

// --------- 以下为user action ---------

func (g *Game) EndGame() {
	g.GameEnd = true
	if g.OperationPlayer != nil {
		g.TurnChan <- true
	}
}

func (g *Game) Born() {
	area := g.PlayerAreas[g.OperationPlayer.Session.ID]
	var picked []*Card
	picked, area.Egg, _ = ListPick(area.Egg, 1)
	monster := NewMonsterCard()
	monster.ID = g.InstanceIDCounter
	g.InstanceIDCounter++
	monster.List = picked
	area.Born = append(area.Born, monster)
	g.TurnChan <- true
}

func (g *Game) BornSummon() {
	area := g.PlayerAreas[g.OperationPlayer.Session.ID]
	area.Born, area.Field, _ = ListMove(area.Born, area.Field, 1)
	g.TurnChan <- true
}

func (g *Game) Attack(card *MonsterCard, target *MonsterCard) {
}

func (g *Game) Summon(id int) {
	area := g.PlayerAreas[g.OperationPlayer.Session.ID]
	index := ListFindIndex(area.Hand, func(it *Card) bool {
		return it.ID == id
	})
	if index == -1 {
		return
	}
	card := area.Hand[index]
	area.Hand = ListRemoveAt(area.Hand, index)
	monster := NewMonsterCard()
	monster.ID = g.InstanceIDCounter
	g.InstanceIDCounter++
	monster.List = []*Card{card}
	area.Field = append(area.Field, monster)
	g.BroadcastGameInfo()
}

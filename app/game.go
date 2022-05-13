package app

import (
	"encoding/json"
	"errors"
	"log"
	"time"

	"github.com/Mrtoy/dtcg-server/service"
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
	TurnChan          chan bool             `json:"-"`
	EndChan           chan bool             `json:"-"`
	MessageChan       chan *service.Package `json:"-"`
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
		MessageChan: make(chan *service.Package),
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

	g.setupDeck(g.Players[0])
	g.setupDeck(g.Players[1])
	g.BroadcastGameInfo()

	g.setupDraw(g.Players[0])
	g.setupDraw(g.Players[1])
	g.BroadcastGameInfo()

	g.CurrentTurn = "born"

	for {
		if g.GameEnd {
			g.onGameEnd()
			return
		}
		g.update()
		timer := time.After(233 * time.Second)
		if g.OperationPlayer != nil {
			log.Printf("player %d -- wait user action", g.OperationPlayer.Session.ID)
		}
	loop:
		for {
			if g.OperationPlayer == nil {
				break loop
			}
			select {
			case pack := <-g.MessageChan:
				g.onAction(pack)
			case <-g.TurnChan:
				break loop
			case <-timer:
				break loop
			case <-g.EndChan:
				g.GameEnd = true
				break loop
			}
		}
		g.OperationPlayer = nil
	}
}

func (g *Game) update() {
	switch g.CurrentTurn {
	case "active":
		g.onTurnActive()
	case "draw":
		g.onTurnDraw()
	case "born":
		g.onTurnBorn()
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

func (g *Game) setupDeck(p *Player) {
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

func (g *Game) setupDraw(player *Player) {
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

func (g *Game) onTurnActive() {

}

func (g *Game) onTurnDraw() {
	area := g.PlayerAreas[g.CurrentPlayer.Session.ID]
	var err error
	area.Deck, area.Hand, err = ListMove(area.Deck, area.Hand, 1)
	if err != nil {
		g.WinPlayer = g.CurrentPlayer.Opponent
		g.GameEnd = true
		return
	}
}

func (g *Game) onTurnBorn() {
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
}

func (g *Game) summon(id int) {
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

}

func (g *Game) onAction(pack *service.Package) {
	log.Printf("player %d -- user action %s", g.OperationPlayer.Session.ID, string(pack.Type))
	switch string(pack.Type) {
	case "game:play-card":
		g.onActionPlayCard(pack)
	case "game:born":
		g.onActionBorn()
	case "game:attack":
		g.onActionAttack()
	case "game:next-turn":
		g.OperationPlayer = nil
	}
	g.BroadcastGameInfo()
}

func (g *Game) onActionPlayCard(pack *service.Package) {
	var req struct {
		ID int `json:"id"`
	}
	pack.Unmarshal(&req)
	g.summon(req.ID)
}

func (g *Game) onActionBorn() {
	area := g.PlayerAreas[g.OperationPlayer.Session.ID]
	if len(area.Born) == 0 {
		var picked []*Card
		picked, area.Egg, _ = ListPick(area.Egg, 1)
		monster := NewMonsterCard()
		monster.ID = g.InstanceIDCounter
		g.InstanceIDCounter++
		monster.List = picked
		area.Born = append(area.Born, monster)
	} else {
		area.Born, area.Field, _ = ListMove(area.Born, area.Field, 1)
	}
	g.OperationPlayer = nil
}

func (g *Game) onActionAttack() {
}

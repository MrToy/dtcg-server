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
	Field2  []*MonsterCard //驯兽师场
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
		Field2:  []*MonsterCard{},
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

func (g *Game) broadcastGameInfo() {
	dmp := diffmatchpatch.New()
	bytes, _ := json.Marshal(g)
	patchs := dmp.PatchMake(g.PreStateStr, string(bytes))
	g.PreStateStr = string(bytes)
	g.Broadcast("game:info-diff", dmp.PatchToText(patchs))
}

func (g *Game) broadcastDeckDetails() {
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
	g.broadcastDeckDetails()

	g.CurrentPlayer = g.Players[0]

	g.setupDeck(g.Players[0])
	g.setupDeck(g.Players[1])
	g.broadcastGameInfo()

	g.setupDraw(g.Players[0])
	g.setupDraw(g.Players[1])
	g.broadcastGameInfo()

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
		g.onTurnMain()
	case "end":
		g.onTurnEnd()
	}
	g.broadcastGameInfo()

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
	g.broadcastGameInfo()
	g.Broadcast("game:end", nil)
	for _, p := range g.Players {
		delete(p.Session.Data, "game")
	}
}

func (g *Game) setupDeck(p *Player) {
	deck := []*Card{}
	for _, s := range p.OriginDeck {
		card := NewCard(g)
		card.Serial = s
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
	area := g.PlayerAreas[g.CurrentPlayer.Session.ID]
	if len(area.Egg) == 0 {
		return
	}
	message := "是否需要育成?"
	if len(area.Born) > 0 {
		message = "是否登场育成区怪兽?"
	}
	g.CurrentPlayer.Session.Send("confirm", map[string]any{
		"message":  message,
		"callback": "game:born",
	})
	g.OperationPlayer = g.CurrentPlayer
}

func (g *Game) onTurnMain() {
	g.OperationPlayer = g.CurrentPlayer
}

func (g *Game) onTurnEnd() {
	g.CurrentPlayer = g.CurrentPlayer.Opponent
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
	g.broadcastGameInfo()
}

func (g *Game) onActionPlayCard(pack *service.Package) {
	var req struct {
		ID int `json:"id"`
	}
	pack.Unmarshal(&req)
	area := g.PlayerAreas[g.OperationPlayer.Session.ID]
	index := ListFindIndex(area.Hand, func(it *Card) bool {
		return it.ID == req.ID
	})
	if index == -1 {
		return
	}
	card := area.Hand[index]
	d := GetDetail(card.Serial)
	area.Hand = ListRemove(area.Hand, card)
	switch d.Type {
	case "选项卡":
		g.playMagicCard(card)
	case "数码兽卡":
		g.summon(card)
	case "驯兽师卡":
		g.summonTamer(card)
	}
}

func (g *Game) playMagicCard(card *Card) {
	area := g.PlayerAreas[g.OperationPlayer.Session.ID]
	area.Discard = append(area.Discard, card)
}

func (g *Game) summon(card *Card) {
	area := g.PlayerAreas[g.OperationPlayer.Session.ID]
	monster := NewMonsterCard(g)
	monster.List = []*Card{card}
	area.Field = append(area.Field, monster)
}

func (g *Game) summonTamer(card *Card) {
	area := g.PlayerAreas[g.OperationPlayer.Session.ID]
	monster := NewMonsterCard(g)
	monster.List = []*Card{card}
	area.Field2 = append(area.Field2, monster)
}

func (g *Game) onActionBorn() {
	area := g.PlayerAreas[g.OperationPlayer.Session.ID]
	if len(area.Born) == 0 {
		g.born()
	} else {
		g.summonBorn()
	}
	g.OperationPlayer = nil
}

func (g *Game) born() {
	area := g.PlayerAreas[g.OperationPlayer.Session.ID]
	var picked []*Card
	picked, area.Egg, _ = ListPick(area.Egg, 1)
	monster := NewMonsterCard(g)
	monster.List = picked
	area.Born = append(area.Born, monster)
}

func (g *Game) summonBorn() {
	area := g.PlayerAreas[g.OperationPlayer.Session.ID]
	area.Born, area.Field, _ = ListMove(area.Born, area.Field, 1)
}

func (g *Game) onActionAttack() {

}

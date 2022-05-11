package app

import (
	"encoding/json"
	"errors"
	"time"

	"github.com/sergi/go-diff/diffmatchpatch"
)

var ErrInvalidOperation = errors.New("invalid operation")

type PlayerArea struct {
	Egg     []*Card        //数码蛋
	Born    *MonsterCard   //育成区
	Hand    []*Card        //手牌
	Field   []*MonsterCard //场上
	Deck    []*Card        //卡组
	Discard []*Card        //弃牌堆
	Defense []*Card        //安防区
}

type Game struct {
	*Room            `json:"-"`
	PlayerAreas      map[int]*PlayerArea
	MemoryBank       int //内存条
	CurrentPlayer    *Player
	OperationPlayer  *Player
	TurnChan         chan bool `json:"-"`
	EndChan          chan bool `json:"-"`
	WinPlayer        *Player
	PreStateStr      string `json:"-"`
	CurrentTurn      string
	CurrentTurnIndex int
}

func NewGame(room *Room) *Game {
	g := &Game{
		Room:        room,
		MemoryBank:  0,
		TurnChan:    make(chan bool),
		EndChan:     make(chan bool),
		PlayerAreas: make(map[int]*PlayerArea),
	}
	for _, player := range room.Players {
		area := &PlayerArea{}
		g.PlayerAreas[player.Session.ID] = area
	}
	g.Players[0].Opponent = g.Players[1]
	g.Players[1].Opponent = g.Players[0]
	return g
}

func (g *Game) BroadcastGameInfo() {
	dmp := diffmatchpatch.New()
	bytes, _ := json.Marshal(g)
	patchs := dmp.PatchMake(g.PreStateStr, string(bytes))
	g.PreStateStr = string(bytes)
	g.Broadcast("game:info-diff", dmp.PatchToText(patchs))
	// log.Println(dmp.PatchApply(patchs, g.PreStateStr))
}

func (g *Game) Start() {
	g.Broadcast("game:start", nil)

	g.SetupPlayer(g.Players[0])
	g.SetupPlayer(g.Players[1])

	g.CurrentPlayer = g.Players[0]
	g.OperationPlayer = g.CurrentPlayer
	g.BroadcastGameInfo()

	for {
		g.OnTurnChange()
		select {
		case <-g.EndChan:
			return
		case <-g.TurnChan:
		case <-time.After(30 * time.Second):
		}
	}
}

func (g *Game) End() {
	g.Broadcast("game:end", nil)
	g.EndChan <- true
}

func (g *Game) SetupPlayer(player *Player) {
	g.SpliteOriginDeck(player)
	area := g.PlayerAreas[player.Session.ID]
	ShuffleCards(area.Egg)
	ShuffleCards(area.Deck)

	area.Deck, area.Defense, _ = ListMove(area.Deck, area.Defense, 5)
	area.Deck, area.Hand, _ = ListMove(area.Deck, area.Hand, 5)
}

func (g *Game) SpliteOriginDeck(p *Player) {
	mainDeck := []*Card{}
	eggDeck := []*Card{}
	for _, s := range p.OriginDeck {
		d := GetDetail(s)
		if d.Level == "2" {
			eggDeck = append(eggDeck, NewCard(s))
		} else {
			mainDeck = append(mainDeck, NewCard(s))
		}
	}
	g.PlayerAreas[p.Session.ID].Egg = eggDeck
	g.PlayerAreas[p.Session.ID].Deck = mainDeck
}

var turnList = []string{"active", "draw", "born", "main", "end"}

func (g *Game) OnTurnChange() {
	g.CurrentTurnIndex++
	if g.CurrentTurnIndex > len(turnList)-1 {
		g.CurrentTurnIndex = 0
	}
	g.CurrentTurn = turnList[g.CurrentTurnIndex]
	g.BroadcastGameInfo()
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
}

func (g *Game) OnTurnActive() {

	g.TurnChan <- true
}

func (g *Game) OnTurnDraw() {
	area := g.PlayerAreas[g.CurrentPlayer.Session.ID]
	var err error
	area.Deck, area.Hand, err = ListMove(area.Deck, area.Hand, 1)
	if err != nil {
		g.WinPlayer = g.CurrentPlayer.Opponent
		g.End()
		return
	}
	g.TurnChan <- true
}

func (g *Game) OnTurnBorn() {
	g.CurrentPlayer.Session.Send("confirm", map[string]any{
		"message":  "是否需要育成",
		"callback": "game:born",
	})
}

func (g *Game) OnTurnMain() {

}

func (g *Game) OnTurnEnd() {
	g.CurrentPlayer = g.CurrentPlayer.Opponent
	g.OperationPlayer = g.CurrentPlayer
	g.BroadcastGameInfo()
	g.TurnChan <- true
}

func (g *Game) Born() {
	area := g.PlayerAreas[g.OperationPlayer.Session.ID]
	var picked []*Card
	picked, area.Egg, _ = ListPick(area.Egg, 1)
	monster := NewMonsterCard()
	monster.List = picked
	area.Born = monster
	g.TurnChan <- true
}

func (g *Game) SummonBorn() {
	area := g.PlayerAreas[g.OperationPlayer.Session.ID]
	if area.Born == nil {
		return
	}
	area.Field = append(area.Field, area.Born)
	area.Born = nil
	g.TurnChan <- true
}

func (g *Game) Summon(id int) {
	area := g.PlayerAreas[g.OperationPlayer.Session.ID]
	if area.Born == nil {
		return
	}
	index := ListFindIndex(area.Hand, func(it *Card) bool {
		return it.ID == id
	})
	if index == -1 {
		return
	}
	card := area.Hand[index]
	area.Hand = ListRemoveAt(area.Hand, index)
	monster := NewMonsterCard()
	monster.List = []*Card{card}
	area.Field = append(area.Field, monster)
}

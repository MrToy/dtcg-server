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
	gameEnd           bool
	TurnCount         int
	EachTurnEvoCount  int //本回合进化过的次数
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

func (g *Game) broadcast(tp string, data any) {
	for _, p := range g.Players {
		err := p.Session.Send(tp, data)
		if err != nil {
			log.Println("send error:", err)
		}
	}
}

func (g *Game) broadcastGameInfo() {
	if g.gameEnd {
		return
	}
	dmp := diffmatchpatch.New()
	bytes, _ := json.Marshal(g)
	patchs := dmp.PatchMake(g.PreStateStr, string(bytes))
	g.PreStateStr = string(bytes)
	g.broadcast("game:info-diff", dmp.PatchToText(patchs))
}

func (g *Game) broadcastDeckDetails() {
	details := map[string]*CardDetail{}
	for _, p := range g.Players {
		for _, s := range p.OriginDeck {
			details[s] = GetDetail(s)
		}
	}
	g.broadcast("game:deck-details", details)
}

type TurnChain struct {
	Action func()
	Next   *TurnChain
}

func (g *Game) Start() {
	g.broadcast("game:start", nil)
	g.CurrentPlayer = g.Players[0]
	g.setupGame()

	currentChain := g.getFirstTurn()
	for {
		currentChain.Action()
		g.processEffect("", nil)
		currentChain = currentChain.Next
		if g.gameEnd {
			break
		}
	}
	g.onGameEnd()
}

func (g *Game) onGameEnd() {
	g.broadcastGameInfo()
	g.broadcast("game:end", nil)
	for _, p := range g.Players {
		delete(p.Session.Data, "game")
	}
}

func (g *Game) setupGame() {
	g.broadcastDeckDetails()
	g.setupDeck(g.Players[0])
	g.setupDeck(g.Players[1])
	g.broadcastGameInfo()

	g.setupDraw(g.Players[0])
	g.setupDraw(g.Players[1])
	g.broadcastGameInfo()
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

func (g *Game) getFirstTurn() *TurnChain {
	turnEnd := &TurnChain{
		Action: g.onTurnEnd,
	}
	turnMain := &TurnChain{
		Action: g.onTurnMain,
		Next:   turnEnd,
	}
	turnBorn := &TurnChain{
		Action: g.onTurnBorn,
		Next:   turnMain,
	}
	turnDraw := &TurnChain{
		Action: g.onTurnDraw,
		Next:   turnBorn,
	}
	turnActive := &TurnChain{
		Action: g.onTurnActive,
		Next:   turnDraw,
	}
	turnEnd.Next = turnActive
	return turnBorn
}

func (g *Game) onTurnActive() {
	area := g.PlayerAreas[g.CurrentPlayer.Session.ID]
	for _, monster := range area.Field {
		monster.Sleep = false
	}
	for _, monster := range area.Field2 {
		monster.Sleep = false
	}
	g.broadcastGameInfo()
}

func (g *Game) onTurnDraw() {
	g.Draw(g.CurrentPlayer, 1)
	g.broadcastGameInfo()
}

func (g *Game) Draw(p *Player, n int) {
	area := g.PlayerAreas[p.Session.ID]
	var err error
	area.Deck, area.Hand, err = ListMove(area.Deck, area.Hand, n)
	if err != nil {
		g.WinPlayer = g.CurrentPlayer.Opponent
		g.gameEnd = true
	}
}

func (g *Game) onTurnBorn() {
	area := g.PlayerAreas[g.CurrentPlayer.Session.ID]
	if len(area.Egg) == 0 {
		return
	}
	content := "是否需要育成?"
	if len(area.Born) > 0 {
		content = "是否登场育成区怪兽?"
	}
	g.CurrentPlayer.Session.Send("confirm", map[string]any{
		"title":    "育成阶段",
		"content":  content,
		"callback": "game:born",
	})
	g.waitForAction(g.CurrentPlayer)
}

func (g *Game) onTurnMain() {
	for {
		nextTurn := g.waitForAction(g.CurrentPlayer)
		if nextTurn {
			break
		}
	}
}

func (g *Game) onTurnEnd() {
	g.TurnCount++
	g.EachTurnEvoCount = 0
	g.CurrentPlayer = g.CurrentPlayer.Opponent
}

func (g *Game) waitForAction(p *Player) (nextTurn bool) {
	if g.gameEnd {
		return
	}
	g.OperationPlayer = p
	log.Printf("player %d -- wait for action", p.Session.ID)
	select {
	case pack := <-g.MessageChan:
		nextTurn = g.onAction(pack)
	case <-g.EndChan:
		g.gameEnd = true
	case <-time.After(233 * time.Second):
	}
	g.OperationPlayer = nil
	return
}

func (g *Game) onAction(pack *service.Package) (nextTurn bool) {
	switch string(pack.Type) {
	case "game:play-card":
		g.onActionPlayCard(pack)
	case "game:born":
		g.onActionBorn()
	case "game:attack":
		g.onActionAttack()
	case "game:next-turn":
		nextTurn = true
	}
	g.broadcastGameInfo()
	return
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
	monster.Add(card)
	area.Field = append(area.Field, monster)
	g.processEffect("summon", monster)
}

func (g *Game) evo(card *Card) {
	area := g.PlayerAreas[g.OperationPlayer.Session.ID]
	monster := NewMonsterCard(g)
	monster.Add(card)
	area.Field = append(area.Field, monster)
	g.processEffect("evo", monster)
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
	// g.processEffect("attack")
	g.processDead()
}

func (g *Game) processEffect(activeTime string, triggerMonster *MonsterCard) {
	for _, p := range g.Players {
		area := g.PlayerAreas[p.Session.ID]
		for _, monster := range area.Field {
			if triggerMonster != nil && triggerMonster != monster {
				continue
			}
			monster.DP = monster.GetOriginDP()
			for _, card := range monster.List {
				for _, effect := range card.EffectList {
					ctx := &CardEffectContext{
						Game:    g,
						Card:    card,
						Monster: monster,
						Effect:  &effect,
						Belong:  p,
					}
					if effect.ActiveTime == "" || activeTime == effect.ActiveTime {
						effect.Action(ctx)
					}
				}
			}
		}
	}
	g.processDead()
}

func (g *Game) processDead() {
	for _, p := range g.Players {
		area := g.PlayerAreas[p.Session.ID]
		for _, monster := range area.Field {
			if monster.DP < 0 {
				g.processEffect("destroy", monster)
				area.Discard = append(area.Discard, monster.List...)
				area.Field = ListRemove(area.Field, monster)
			}
		}
	}
}

package app

import (
	_ "embed"
	"encoding/json"
	"math/rand"
	"time"
)

//go:embed data2.json
var cardDetailsBuf []byte

var CardDetails map[string]CardDetail

func init() {
	json.Unmarshal(cardDetailsBuf, &CardDetails)
}

type CardDetail struct {
	Serial         string
	Type           string
	Name           string
	Color          []string
	Effect         string
	EvoCoverEffect string
	SecurityEffect string
	Level          int
	Cost           int
	Cost1          int
	DP             int
	Attribute      string
	Class          []string
	Image          string
}

type Card struct {
	ID            int
	Serial        string
	EffectList    []CardEffect
	Cost          int
	OriginCost    int
	EvoCost       int
	OriginEvoCost int
}

func (c *Card) AddEffect(e CardEffect) {
	c.EffectList = append(c.EffectList, e)
}

func NewCard(g *Game) *Card {
	g.InstanceIDCounter++
	return &Card{
		ID:         g.InstanceIDCounter,
		EffectList: []CardEffect{},
	}
}

func ShuffleCards(list []*Card) {
	rand.Seed(time.Now().UnixNano())
	rand.Shuffle(len(list), func(i, j int) { list[i], list[j] = list[j], list[i] })
}

type MonsterCard struct {
	ID       int
	List     []*Card
	Sleep    bool
	DP       int
	OriginDP int
}

func NewMonsterCard(g *Game) *MonsterCard {
	g.InstanceIDCounter++
	return &MonsterCard{
		List: []*Card{},
		ID:   g.InstanceIDCounter,
	}
}

func (m *MonsterCard) Add(c *Card) {
	m.List = append([]*Card{c}, m.List...)
	m.OriginDP = CardDetails[c.Serial].DP
	m.DP = m.OriginDP
}

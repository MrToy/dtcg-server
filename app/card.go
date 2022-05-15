package app

import (
	_ "embed"
	"encoding/json"
	"math/rand"
	"strconv"
	"time"
)

//go:embed response.json
var cardDetailsBuf []byte

var details map[string]*CardDetail

func init() {
	details = GetCardDetails()
}

func GetCardDetails() map[string]*CardDetail {
	var res struct {
		Data struct {
			List []*CardDetail `json:"list"`
		} `json:"data"`
	}
	json.Unmarshal(cardDetailsBuf, &res)
	m := make(map[string]*CardDetail)
	for _, v := range res.Data.List {
		m[v.Serial] = v
	}
	return m
}

type CardDetail struct {
	Serial         string
	Type           string
	Name           string
	Color          []string
	Effect         string
	EvoCoverEffect string
	SecurityEffect string
	Level          string
	Cost           string
	EvoCost        string
	DP             string
	Attribute      string
	Class          []string
	Images         []struct {
		ImgPath   string
		ThumbPath string
	}
}

type Card struct {
	ID         int
	Serial     string
	EffectList []CardEffect
}

func NewCard(g *Game) *Card {
	g.InstanceIDCounter++
	return &Card{
		ID:         g.InstanceIDCounter,
		EffectList: []CardEffect{},
	}
}

func GetDetail(serial string) *CardDetail {
	return details[serial]
}

func ShuffleCards(list []*Card) {
	rand.Seed(time.Now().UnixNano())
	rand.Shuffle(len(list), func(i, j int) { list[i], list[j] = list[j], list[i] })
}

type MonsterCard struct {
	ID    int
	List  []*Card
	Sleep bool
	DP    int
}

func NewMonsterCard(g *Game) *MonsterCard {
	g.InstanceIDCounter++
	return &MonsterCard{
		List: []*Card{},
		ID:   g.InstanceIDCounter,
	}
}

func (m *MonsterCard) GetOriginDP() int {
	dp, _ := strconv.Atoi(GetDetail(m.List[0].Serial).DP)
	return dp
}

func (m *MonsterCard) Add(c *Card) {
	m.List = append([]*Card{c}, m.List...)
	m.DP = m.GetOriginDP()
}

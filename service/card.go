package service

import (
	_ "embed"
	"encoding/json"
	"errors"
	"math/rand"
	"time"
)

//go:embed response.json
var buf []byte

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
	json.Unmarshal(buf, &res)
	m := make(map[string]*CardDetail)
	for _, v := range res.Data.List {
		m[v.Serial] = v
	}
	return m
}

type CardDetail struct {
	Serial         string   `json:"serial"`
	Type           string   `json:"type"`
	Name           string   `json:"name"`
	Color          []string `json:"color"`
	Effect         string   `json:"effect"`
	EvoCoverEffect string   `json:"evo_cover_effect"`
	SecurityEffect string   `json:"security_effect"`
	Level          string   `json:"level"`
	Cost           string   `json:"cost"`
	EvoCost        string   `json:"cost_1"`
	DP             string   `json:"dp"`
	Attribute      string   `json:"attribute"`
	Class          []string `json:"class"`
	Images         []struct {
		ImgPath   string `json:"img_path"`
		ThumbPath string `json:"thumb_path"`
	} `json:"images"`
}

type Card string

func (c Card) GetDetail() *CardDetail {
	return details[string(c)]
}

func ShuffleCards(list []Card) {
	rand.Seed(time.Now().UnixNano())
	rand.Shuffle(len(list), func(i, j int) { list[i], list[j] = list[j], list[i] })
}

type CardMonster struct {
	List  []Card
	Sleep bool
}

func CreateCardMonster(card Card) *CardMonster {
	return &CardMonster{
		List: []Card{card},
	}
}

func MoveCards(src []Card, dist []Card, num int) ([]Card, []Card, error) {
	if len(src) < num {
		return nil, nil, errors.New("not enough cards")
	}
	return src[:len(src)-num], append(dist, src[len(src)-num:]...), nil
}

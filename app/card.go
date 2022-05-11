package app

import (
	_ "embed"
	"encoding/json"
	"errors"
	"math/rand"
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
	ID     int
	Serial string
}

var globalCardID int = 0

func NewCard(s string) *Card {
	globalCardID++
	return &Card{
		ID:     globalCardID,
		Serial: s,
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
}

func NewMonsterCard() *MonsterCard {
	globalCardID++
	return &MonsterCard{
		ID: globalCardID,
	}
}

func ListPick[T comparable](src []T, num int) ([]T, []T, error) {
	if len(src) < num {
		return nil, nil, errors.New("not enough cards")
	}
	return src[len(src)-num:], src[:len(src)-num], nil
}

func ListRemoveAt[T comparable](list []T, index int) []T {
	return append(list[:index], list[index+1:]...)
}

func ListAppend[T comparable](src []T, dst []T) []T {
	return append(dst, src...)
}

func ListMove[T comparable](src []T, dist []T, num int) ([]T, []T, error) {
	if len(src) < num {
		return nil, nil, errors.New("not enough")
	}
	return src[:len(src)-num], append(dist, src[len(src)-num:]...), nil
}

func ListDifference[T comparable](a, b []T) []T {
	mb := make(map[T]struct{}, len(b))
	for _, x := range b {
		mb[x] = struct{}{}
	}
	var diff []T
	for _, x := range a {
		if _, found := mb[x]; !found {
			diff = append(diff, x)
		}
	}
	return diff
}

func ListFindIndex[T comparable](list []T, predicate func(a T) bool) int {
	for i := 0; i < len(list); i++ {
		if predicate(list[i]) {
			return i
		}
	}
	return -1
}

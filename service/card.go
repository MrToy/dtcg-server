package service

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

type CardMonster struct {
	ID    int
	List  []*Card
	Sleep bool
}

func NewCardMonster() *CardMonster {
	globalCardID++
	return &CardMonster{
		ID: globalCardID,
	}
}

func PickList[T comparable](src []T, num int) ([]T, []T, error) {
	if len(src) < num {
		return nil, nil, errors.New("not enough cards")
	}
	return src[num:], src[:len(src)-num], nil
}

func AppendList[T comparable](src []T, dst []T) []T {
	return append(dst, src...)
}

func MoveList[T comparable](src []T, dist []T, num int) ([]T, []T, error) {
	if len(src) < num {
		return nil, nil, errors.New("not enough")
	}
	return src[:len(src)-num], append(dist, src[len(src)-num:]...), nil
}

func DifferenceList[T comparable](a, b []T) []T {
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

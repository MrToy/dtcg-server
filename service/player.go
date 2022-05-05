package service

import "errors"

type Event struct {
	Type         string      `json:"type"`
	Target       *Player     `json:"target"`
	Data         interface{} `json:"data"`
	OriginSource []byte      `json:"-"`
}

type Player struct {
	ID         int
	Name       string
	Ready      bool
	Deck       []Card      `json:"-"`
	EggDeck    []Card      `json:"-"`
	Opponent   *Player     `json:"-"`
	ReadChan   chan *Event `json:"-"`
	WriteChan  chan *Event `json:"-"`
	Room       *Room       `json:"-"`
	PlayerArea *PlayerArea `json:"-"`
}

var playerID int = 0

func NewPlayer() *Player {
	playerID++
	p := &Player{
		ReadChan:  make(chan *Event),
		WriteChan: make(chan *Event),
		ID:        playerID,
	}
	return p
}

func (p *Player) UseDeck(deck []Card) error {
	mainDeck := []Card{}
	eggDeck := []Card{}
	for _, c := range deck {
		d := c.GetDetail()
		if d.Level == "2" {
			eggDeck = append(eggDeck, c)
		} else {
			mainDeck = append(mainDeck, c)
		}
	}
	if len(eggDeck) != 5 {
		return errors.New("egg deck must have 5 cards")
	}
	if len(mainDeck) != 50 {
		return errors.New("main deck must have 50 cards")
	}
	p.EggDeck = eggDeck
	p.Deck = mainDeck
	return nil
}

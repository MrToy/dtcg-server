package service

import (
	"errors"

	"github.com/Mrtoy/dtcg-server/server"
)

type Player struct {
	Session    *server.Session
	Name       string
	OriginDeck []Card
	Deck       []Card
	EggDeck    []Card
	Opponent   *Player
	PlayerArea *PlayerArea
}

func SetPlayer(sess *server.Session) {
	player := &Player{
		Session: sess,
	}
	sess.Data["player"] = player
}

func UpdatePlayerInfo(pack *server.Package, sess *server.Session) {
	player := sess.Data["player"].(*Player)
	var info Player
	err := pack.Unmarshal(&info)
	if err != nil {
		sess.Error(err)
		return
	}
	player.Name = info.Name
	err = player.UseDeck(info.OriginDeck)
	if err != nil {
		sess.Error(err)
		return
	}
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

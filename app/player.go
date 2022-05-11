package app

import "github.com/Mrtoy/dtcg-server/service"

type Player struct {
	Session    *service.Session
	Name       string
	OriginDeck []string
	Opponent   *Player `json:"-"`
}

func SetPlayer(sess *service.Session) {
	player := &Player{
		Session: sess,
	}
	sess.Data["player"] = player
}

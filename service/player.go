package service

import (
	"github.com/Mrtoy/dtcg-server/server"
)

type Player struct {
	Session    *server.Session
	Name       string
	OriginDeck []string
	Opponent   *Player
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
}

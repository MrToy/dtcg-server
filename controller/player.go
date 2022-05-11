package controller

import (
	"github.com/Mrtoy/dtcg-server/app"
	"github.com/Mrtoy/dtcg-server/service"
)

func UpdatePlayerInfo(pack *service.Package, sess *service.Session) {
	player := sess.Data["player"].(*app.Player)
	var info app.Player
	err := pack.Unmarshal(&info)
	if err != nil {
		sess.Error(err)
		return
	}
	player.Name = info.Name
	player.OriginDeck = info.OriginDeck
}

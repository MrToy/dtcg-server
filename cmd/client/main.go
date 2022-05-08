package main

import (
	"encoding/json"
	"log"
	"net"

	"github.com/Mrtoy/dtcg-server/server"
	"github.com/Mrtoy/dtcg-server/service"
)

const deckInfo = `["Exported from http://digimon.card.moe","ST3-07","ST3-07","ST3-07","ST3-12","ST3-12","ST3-12","ST3-12","BT1-063","BT1-063","BT1-052","BT1-052","BT1-052","BT1-052","BT1-048","BT1-048","BT1-048","BT1-048","BT1-005","BT1-006","BT1-006","BT1-006","BT1-006","BT1-057","BT1-057","BT1-102","BT1-102","BT2-033","BT2-033","BT2-033","BT2-099","BT2-038","BT2-038","BT2-038","BT2-038","BT2-034","BT2-034","BT2-034","BT2-039","BT2-039","BT2-098","BT2-098","BT1-060","BT1-060","BT1-060","BT1-060","BT1-087","BT1-087","BT1-087","BT2-041","BT2-041","BT2-041","BT2-041","BT2-087","BT2-087","BT2-087"]`

func main() {
	conn, err := net.Dial("tcp", "127.0.0.1:2333")
	if err != nil {
		panic(err)
	}
	defer conn.Close()
	sess := server.NewSession(conn)
	sess.On("connect", func(pack *server.Package, sess *server.Session) {
		var res struct {
			ID int
		}
		pack.Unmarshal(&res)
		sess.ID = res.ID
		sess.Send("player:update-info", &service.Player{
			Name: "toy",
			Deck: prepareCard(),
		})
		sess.Send("room:join", map[string]any{"RoomID": "1"})
		sess.Send("room:ready", nil)
	})
	sess.On("error", func(pack *server.Package, sess *server.Session) {
		log.Println(pack.String())
	})
	sess.On("confirm", func(pack *server.Package, sess *server.Session) {
		var res map[string]any
		pack.Unmarshal(&res)
		log.Println(res)
		method := res["callback"].(string)
		sess.Send(method, "1")
	})

	sess.On("room:info", func(pack *server.Package, sess *server.Session) {
		var res map[string]any
		pack.Unmarshal(&res)
		log.Println(res)
	})

	sess.On("game:start", func(pack *server.Package, sess *server.Session) {
	})
	sess.On("game:current-change", func(pack *server.Package, sess *server.Session) {
		var res map[string]any
		pack.Unmarshal(&res)
		log.Println(res)
	})
	sess.On("game:turn-change", func(pack *server.Package, sess *server.Session) {
		log.Println(pack.String())
	})
	sess.On("game:end", func(pack *server.Package, sess *server.Session) {
		var res map[string]any
		pack.Unmarshal(&res)
		log.Println(res)
	})

	sess.Receive()
}

func prepareCard() []service.Card {
	var deck []service.Card
	json.Unmarshal([]byte(deckInfo), &deck)
	deck = deck[1:]
	return deck
}

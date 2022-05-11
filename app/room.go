package app

import (
	"log"
)

type Room struct {
	Name       string
	Players    []*Player
	ReadyState map[int]bool
}

func NewRoom() *Room {
	return &Room{
		Players:    []*Player{},
		ReadyState: make(map[int]bool),
	}
}

func (r *Room) Broadcast(tp string, data any) {
	for _, p := range r.Players {
		err := p.Session.Send(tp, data)
		if err != nil {
			log.Println("send error:", err)
		}
	}
}

func (r *Room) BroadcastRoomInfo() {
	r.Broadcast("room:info", r)
}

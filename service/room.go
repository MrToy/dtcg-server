package service

import (
	"encoding/json"
	"errors"
	"log"
)

type Room struct {
	ID      string
	Players []*Player
	Game    *Game
}

func (r *Room) Join(p *Player) error {
	if len(r.Players) >= 2 {
		return errors.New("room is full")
	}
	p.Room = r
	r.Players = append(r.Players, p)
	r.BroadCastRoomUpdate()
	return nil
}

func (r *Room) Leave(p *Player) {
	for i, player := range r.Players {
		if player == p {
			r.Players = append(r.Players[:i], r.Players[i+1:]...)
			if r.Game != nil {
				r.Game.Players = r.Players
				r.BroadCastRoomUpdate()
				r.Game.WinPlayer = p.Opponent
				r.Game.End()
			}
		}
	}
}

func (r *Room) HandleEvent(e *Event) (err error) {
	switch e.Type {
	case "room:ready":
		err = r.Ready(e)
	case "room:leave":
		r.Leave(e.Target)
	}
	return
}

func (r *Room) BroadCastEvent(e *Event) {
	for _, p := range r.Players {
		p.WriteChan <- e
	}
}

func (r *Room) BroadCastRoomUpdate() {
	r.BroadCastEvent(&Event{
		Type: "room:update-user",
		Data: r.Players,
	})
}

func (r *Room) Ready(event *Event) error {
	var req struct {
		Data struct {
			Ready bool   `json:"ready"`
			Deck  []Card `json:"deck"`
		} `json:"data"`
	}
	err := json.Unmarshal(event.OriginSource, &req)
	if err != nil {
		return err
	}
	if req.Data.Ready {
		err := event.Target.UseDeck(req.Data.Deck)
		if err != nil {
			return err
		}
	}
	event.Target.Ready = req.Data.Ready
	readyCount := 0
	for _, p := range event.Target.Room.Players {
		if p.Ready {
			readyCount++
		}
	}
	if readyCount == 2 {
		go r.StartGame()
	}
	return nil
}

func (r *Room) StartGame() {
	log.Println("start game")
	r.Game = NewGame()
	r.Game.Players = r.Players
	r.Game.Start()
}

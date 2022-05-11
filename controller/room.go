package controller

import (
	"errors"

	"github.com/Mrtoy/dtcg-server/app"
	"github.com/Mrtoy/dtcg-server/service"
)

type RoomManager struct {
	RoomPlayerLimit int
	Rooms           map[string]*app.Room
	PlayerInRoom    map[int]*app.Room
	GameManager     *GameManager
}

func NewRoomManager() *RoomManager {
	return &RoomManager{
		Rooms:           make(map[string]*app.Room),
		PlayerInRoom:    make(map[int]*app.Room),
		RoomPlayerLimit: 2,
	}
}

func (r *RoomManager) Join(pack *service.Package, sess *service.Session) {
	player := sess.Data["player"].(*app.Player)
	var req struct {
		Name string
	}
	err := pack.Unmarshal(&req)
	if err != nil {
		sess.Error(err)
		return
	}
	room := r.Rooms[req.Name]
	if room == nil {
		room = app.NewRoom()
		room.Name = req.Name
		r.Rooms[req.Name] = room
	}
	if len(room.Players) >= r.RoomPlayerLimit {
		sess.Error(errors.New("room is full"))
		return
	}
	r.PlayerInRoom[player.Session.ID] = room
	room.Players = append(room.Players, player)
	room.BroadcastRoomInfo()
}

func (r *RoomManager) Leave(pack *service.Package, sess *service.Session) {
	player := sess.Data["player"].(*app.Player)
	room, ok := r.PlayerInRoom[player.Session.ID]
	if !ok {
		return
	}
	room.ReadyState[player.Session.ID] = false
	delete(r.PlayerInRoom, player.Session.ID)
	for i, p := range room.Players {
		if player == p {
			room.Players = append(room.Players[:i], room.Players[i+1:]...)
		}
	}
	room.BroadcastRoomInfo()
}

func (r *RoomManager) Ready(pack *service.Package, sess *service.Session) {
	player := sess.Data["player"].(*app.Player)
	room := r.PlayerInRoom[player.Session.ID]
	if room == nil {
		sess.Error(errors.New("you are not in a room"))
		return
	}
	readyCount := 0
	room.ReadyState[player.Session.ID] = true
	for _, ready := range room.ReadyState {
		if ready {
			readyCount++
		}
	}
	room.BroadcastRoomInfo()
	if readyCount == r.RoomPlayerLimit {
		r.GameManager.StartGame(room)
	}
}

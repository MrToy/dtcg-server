package controller

import (
	"errors"

	"github.com/Mrtoy/dtcg-server/app"
	"github.com/Mrtoy/dtcg-server/service"
)

type RoomController struct {
	RoomPlayerLimit int
	Rooms           map[string]*app.Room
	SoloPlayers     map[int]*app.Player
}

func (m *RoomController) JoinSolo(pack *service.Package, sess *service.Session) {
	player := sess.Data["player"].(*app.Player)

	if len(m.SoloPlayers) > 0 {
		for id, p := range m.SoloPlayers {
			delete(m.SoloPlayers, id)
			StartGame([]*app.Player{p, player})
			return
		}
	}
	m.SoloPlayers[sess.ID] = player
}

func (m *RoomController) LeaveSolo(pack *service.Package, sess *service.Session) {
	delete(m.SoloPlayers, sess.ID)
}

func NewRoomController() *RoomController {
	return &RoomController{
		Rooms:           make(map[string]*app.Room),
		RoomPlayerLimit: 2,
		SoloPlayers:     make(map[int]*app.Player),
	}
}

func (r *RoomController) Join(pack *service.Package, sess *service.Session) {
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
	sess.Data["room"] = room
	room.Players = append(room.Players, player)
	room.BroadcastRoomInfo()
}

func (r *RoomController) Leave(pack *service.Package, sess *service.Session) {
	player := sess.Data["player"].(*app.Player)
	room, ok := sess.Data["room"].(*app.Room)
	if !ok {
		return
	}
	delete(room.ReadyState, player.Session.ID)
	room.Players = app.ListRemove(room.Players, player)
	room.BroadcastRoomInfo()
}

func (r *RoomController) Ready(pack *service.Package, sess *service.Session) {
	player := sess.Data["player"].(*app.Player)
	room, ok := sess.Data["room"].(*app.Room)
	if !ok {
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
		StartGame(room.Players)
	}
}

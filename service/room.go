package service

import (
	"errors"

	"github.com/Mrtoy/dtcg-server/server"
)

type RoomManager struct {
	RoomPlayerLimit int
	Rooms           map[string]*Room
	PlayerInRoom    map[int]*Room
	GameManager     *GameManager
}

func NewRoomManager() *RoomManager {
	return &RoomManager{
		Rooms:           make(map[string]*Room),
		PlayerInRoom:    make(map[int]*Room),
		RoomPlayerLimit: 2,
	}
}

type Room struct {
	ID         string
	Players    []*Player
	ReadyState map[int]bool
}

func NewRoom() *Room {
	return &Room{
		Players:    []*Player{},
		ReadyState: make(map[int]bool),
	}
}

func (r *RoomManager) Join(pack *server.Package, sess *server.Session) {
	player := sess.Data["player"].(*Player)
	var req struct {
		RoomID string
	}
	err := pack.Unmarshal(&req)
	if err != nil {
		sess.Error(err)
		return
	}
	room := r.Rooms[req.RoomID]
	if room == nil {
		room = NewRoom()
		r.Rooms[req.RoomID] = room
	}
	if len(room.Players) >= r.RoomPlayerLimit {
		sess.Error(errors.New("room is full"))
		return
	}
	r.PlayerInRoom[player.Session.ID] = room
	room.Players = append(room.Players, player)
	room.BroadCastRoomUpdate()
}

func (r *RoomManager) Leave(pack *server.Package, sess *server.Session) {
	player := sess.Data["player"].(*Player)
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
	room.BroadCastRoomUpdate()
}

func (r *RoomManager) Ready(pack *server.Package, sess *server.Session) {
	player := sess.Data["player"].(*Player)
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
	room.BroadCastRoomUpdate()
	if readyCount == r.RoomPlayerLimit {
		r.GameManager.StartGame(room)
	}
}

func (r *Room) BroadCast(tp string, data any) {
	for _, p := range r.Players {
		p.Session.Send(tp, data)
	}
}

func (r *Room) BroadCastRoomUpdate() {
	type UserInfo struct {
		Name  string
		ID    int
		Ready bool
	}
	var players []*UserInfo
	for _, p := range r.Players {
		players = append(players, &UserInfo{
			Name:  p.Name,
			ID:    p.Session.ID,
			Ready: r.ReadyState[p.Session.ID],
		})
	}
	r.BroadCast("room:info", map[string]any{
		"id":      r.ID,
		"players": players,
	})
}

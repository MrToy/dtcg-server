package service

import (
	"bufio"
	"encoding/json"
	"io"
	"log"
	"net"
	"strings"
)

type Server struct {
	Rooms   map[string]*Room
	Players map[int]*Player
}

func NewServer() *Server {
	return &Server{
		Rooms:   map[string]*Room{},
		Players: map[int]*Player{},
	}
}

func (s *Server) AddPlayer() *Player {
	p := NewPlayer()
	s.Players[p.ID] = p
	log.Println("player enter", p.ID)
	return p
}

func (s *Server) RemovePlayer(p *Player) {
	s.Players[p.ID] = nil
	if p.Room != nil {
		p.Room.Leave(p)
	}
	log.Println("player leave", p.ID)
}

func (s *Server) HandleConnect(c net.Conn) {
	p := s.AddPlayer()
	go func() {
		for {
			event := <-p.WriteChan
			if event.Target == nil {
				log.Println("event", event.Type, "sys", "->", p.ID)
			} else {
				log.Println("event", event.Type, event.Target.ID, "->", p.ID)
			}
			buf, err := json.Marshal(event)
			if err != nil {
				log.Println("Marshal failed,err:", err)
				continue
			}
			_, err = c.Write(buf)
			if err != nil {
				log.Println("Write failed,err:", err)
				continue
			}
			c.Write([]byte("\n"))
		}
	}()
	reader := bufio.NewReader(c)
	for {
		buf, err := reader.ReadSlice('\n')
		if err == io.EOF {
			log.Println("client closed")
			s.RemovePlayer(p)
			return
		}
		if err != nil {
			log.Println("read failed,err:", err)
			continue
		}
		var event Event
		err = json.Unmarshal(buf, &event)
		if err != nil {
			log.Println("Unmarshal failed,err:", err, string(buf))
			continue
		}
		event.Target = p
		event.OriginSource = buf
		err = s.HandleEvent(&event)
		if err != nil {
			log.Println("HandleEvent failed,err:", err)
			p.WriteChan <- &Event{
				Type: "error",
				Data: err.Error(),
			}
		}
	}
}

func (s *Server) HandleEvent(event *Event) (err error) {
	if event.Type == "room:join" {
		err = s.JoinRoom(event)
	} else if strings.HasPrefix(event.Type, "room:") {
		if event.Target.Room != nil {
			err = event.Target.Room.HandleEvent(event)
		}
	} else if strings.HasPrefix(event.Type, "game:") {
		if event.Target.Room != nil && event.Target.Room.Game != nil {
			err = event.Target.Room.Game.HandleEvent(event)
		}
	}
	return
}

func (s *Server) JoinRoom(event *Event) error {
	roomID, _ := event.Data.(string)
	if s.Rooms[roomID] == nil {
		s.Rooms[roomID] = &Room{
			ID: roomID,
		}
	}
	s.Rooms[roomID].Join(event.Target)
	return nil
}

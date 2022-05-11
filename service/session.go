package service

import (
	"log"
	"net"
)

type SessionHandler func(pack *Package, sess *Session)

type Session struct {
	ID       int
	conn     net.Conn                  `json:"-"`
	Handlers map[string]SessionHandler `json:"-"`
	Data     map[string]any            `json:"-"`
}

var currentSessionID int = 0

func NewSession(conn net.Conn) *Session {
	currentSessionID++
	return &Session{
		ID:       currentSessionID,
		conn:     conn,
		Handlers: make(map[string]SessionHandler),
		Data:     make(map[string]any),
	}
}

func (s *Session) Send(tp string, data any) error {
	pack, err := NewPackage(tp, data)
	if err != nil {
		return err
	}
	err = pack.Pack(s.conn)
	if err != nil {
		return err
	}
	if string(pack.Type) == "error" {
		log.Printf("player %d <- %s %s", s.ID, pack.Type, pack.Data)
	} else {
		log.Printf("player %d <- %s", s.ID, pack.Type)
	}
	return nil
}

func (s *Session) Error(err error) {
	s.Send("error", err.Error())
}

func (s *Session) On(tp string, handler SessionHandler) {
	s.Handlers[tp] = handler
}

func (s *Session) Receive() {
	for {
		pack := &Package{}
		err := pack.Unpack(s.conn)
		if err != nil {
			log.Printf("player %d -- network error: %s", s.ID, err)
			break
		}
		log.Printf("player %d -> %s", s.ID, pack.Type)
		handler, ok := s.Handlers[string(pack.Type)]
		if ok {
			handler(pack, s)
		}
	}
}

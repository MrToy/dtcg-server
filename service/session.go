package service

import (
	"log"
	"net"
	"path"
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
	log.Printf("player %d <- %s", s.ID, pack.Type)
	return nil
}

func (s *Session) Error(err error) {
	s.Send("error", err.Error())
}

func (s *Session) On(tp string, handler SessionHandler) {
	s.Handlers[tp] = handler
}

func (s *Session) Receive() {
	defer func() {
		if err := recover(); err != nil {
			log.Println("Recovered in f", err)
			s.conn.Close()
		}
	}()
	for {
		pack := &Package{}
		err := pack.Unpack(s.conn)
		if err != nil {
			log.Printf("player %d -- network error: %s", s.ID, err)
			break
		}
		log.Printf("player %d -> %s", s.ID, pack.Type)
		for k, handler := range s.Handlers {
			matched, _ := path.Match(k, string(pack.Type))
			if matched {
				go handler(pack, s)
			}
		}
	}
}

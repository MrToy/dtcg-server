package server

import (
	"log"
	"net"
)

type Server struct {
	Handlers map[string]SessionHandler
}

func NewServer() *Server {
	return &Server{
		Handlers: make(map[string]SessionHandler),
	}
}

func (s *Server) On(tp string, handler SessionHandler) {
	s.Handlers[tp] = handler
}

func (s *Server) Listen(address string) error {
	listener, err := net.Listen("tcp", address)
	if err != nil {
		return err
	}
	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Println("Listen.Accept failed,err:", err)
			continue
		}
		go func(conn net.Conn) {
			sess := NewSession(conn)
			if hander, ok := s.Handlers["connect"]; ok {
				log.Printf("player %d -- %s", sess.ID, "connect")
				hander(nil, sess)
			}
			sess.Handlers = s.Handlers
			sess.Receive()
			if hander, ok := s.Handlers["disconnect"]; ok {
				log.Printf("player %d -- %s", sess.ID, "disconnect")
				hander(nil, sess)
			}
		}(conn)
	}
}

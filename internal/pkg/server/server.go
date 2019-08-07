package server

import (
	"net"

	"github.com/golang/glog"
	"github.com/inhzus/go-redis-impl/internal/pkg/command"
)

type Server struct {
	Proto    string
	Addr     string
	Channels []chan string
}

func (s *Server) Serve() error {
	addr := s.Addr
	if s.Proto == "" {
		s.Proto = "tcp"
	}
	if s.Proto == "unix" && addr == "" {
		addr = "/tmp/redis.sock"
	} else if addr == "" {
		addr = ":6389"
	}
	listener, err := net.Listen(s.Proto, addr)
	if err != nil {
		return err
	}
	s.Channels = []chan string{}
	for {
		conn, err := listener.Accept()
		if err != nil {
			return err
		}
		go func(conn net.Conn) {
			_, err = command.Parse(conn)
			if err != nil {
				glog.Errorf("parse: %v", err)
			}
			_ = conn.Close()
		}(conn)
	}
}

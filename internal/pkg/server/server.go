package server

import (
	"net"

	"github.com/golang/glog"
	"github.com/inhzus/go-redis-impl/internal/pkg/token"
)

type Option struct {
	Proto   string
	Addr    string
	Timeout int64 // millisecond
}

type Server struct {
	option *Option
	queue  chan net.Conn
	stop   chan struct{}
}

func NewServer(option *Option) *Server {

	if option.Proto == "" {
		option.Proto = "tcp"
	}
	if option.Proto == "unix" && option.Addr == "" {
		option.Addr = "/tmp/redis.sock"
	} else if option.Addr == "" {
		option.Addr = ":6389"
	}
	return &Server{option: option,}
}

func (s *Server) Serve() error {
	listener, err := net.Listen(s.option.Proto, s.option.Addr)
	if err != nil {
		return err
	}
	s.queue = make(chan net.Conn)
	s.stop = make(chan struct{})

	go func() {
		for {
			select {
			case <-s.stop:
				glog.Infof("server consumer stop")
				break
			case conn := <-s.queue:
				_, err := token.Deserialize(conn)
				if err != nil {
					glog.Error(err)
				}

				//_ = conn.Close()
			}
		}
	}()

	for {
		select {
		case <-s.stop:
			glog.Infof("server producer stop")
			return listener.Close()
		default:
			conn, err := listener.Accept()
			if err != nil {
				return err
			}
			s.queue <- conn
		}
	}
}

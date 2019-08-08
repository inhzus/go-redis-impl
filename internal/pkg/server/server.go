package server

import (
	"io"
	"net"

	"github.com/golang/glog"
	"github.com/inhzus/go-redis-impl/internal/pkg/command"
	"github.com/inhzus/go-redis-impl/internal/pkg/token"
)

type Option struct {
	Proto   string
	Addr    string
}

type Server struct {
	option *Option
	queue  chan *Task
	stop   chan struct{}
}

type Task struct {
	req *token.Token
	rsp chan *token.Token
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
	return &Server{option: option}
}

func (s *Server) Serve() error {
	listener, err := net.Listen(s.option.Proto, s.option.Addr)
	if err != nil {
		return err
	}
	s.queue = make(chan *Task)
	s.stop = make(chan struct{})

	go func() {
		for {
			select {
			case <-s.stop:
				_ = listener.Close()
				glog.Infof("server closed")
				s.stop <- struct{}{}
				return
			case t := <-s.queue:
				t.rsp <- command.Process(t.req)
			}
		}
	}()

	for {
		conn, err := listener.Accept()
		if err != nil {
			return err
		}
		go func(conn net.Conn) {
			for {
				req, err := token.Deserialize(conn)
				if err != nil {
					if err == io.EOF {
						return
					}
					glog.Error(err)
				}
				glog.Infof("request: %v", req.Format())
				c := make(chan *token.Token)
				s.queue <- &Task{req, c}
				reply := <-c
				rsp, err := reply.Serialize()
				if err != nil {
					glog.Error(err)
					errRsp, _ := token.ErrorDefault.Serialize()
					_, err = conn.Write(errRsp)
				} else {
					_, err = conn.Write(rsp)
				}
				if err != nil {
					glog.Error(err)
				}
			}
		}(conn)
	}
}

func (s *Server) Close() {
	s.stop <- struct{}{}
	<-s.stop
}

package server

import (
	"io"
	"net"

	"github.com/golang/glog"
	"github.com/inhzus/go-redis-impl/internal/pkg/command"
	"github.com/inhzus/go-redis-impl/internal/pkg/label"
	"github.com/inhzus/go-redis-impl/internal/pkg/token"
)

type Option struct {
	Proto   string
	Addr    string
}

type Server struct {
	option *Option
	queue  chan *token.Task
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
	return &Server{option: option}
}

func (s *Server) submit(t *token.Token, conn net.Conn) *token.Token {
	if t == nil {
		return token.NewError("empty request")
	}
	data := t.Data.([]*token.Token)
	if len(data) == 0 {
		return token.NewError("empty token")
	}
	c := make(chan *token.Response)
	if data[0].Label == label.String {
		s.queue <- &token.Task{Req: t, Rsp: c}
		reply := <-c
		return reply.Data
	}
	var rspData []*token.Token
	for _, item := range data {
		rspData = append(rspData, s.submit(item, conn))
	}
	return token.NewArray(rspData...)
}

func (s *Server) handleConnection(conn net.Conn) {
	glog.Infof("client %v connection established", conn.RemoteAddr())
	for {
		req := token.Deserialize(conn)
		if req.Err != nil {
			if req.Err == io.EOF {
				glog.Infof("client %v connection closed", conn.RemoteAddr())
				return
			}
			glog.Error(req.Err)
		}
		glog.Infof("request: %v", req.Data.Format())
		reply := s.submit(req.Data, conn)
		rsp, err := reply.Serialize()
		if err != nil {
			glog.Error(err)
			rsp, _ = token.ErrorDefault.Serialize()
		}
		go func(r []byte) {
			_, err = conn.Write(r)
			if err != nil {
				glog.Error(err)
			}
		}(rsp)
	}
}

func (s *Server) Serve() error {
	listener, err := net.Listen(s.option.Proto, s.option.Addr)
	if err != nil {
		return err
	}
	s.queue = make(chan *token.Task)
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
				t.Rsp <- &token.Response{Data: command.Process(t.Req)}
			}
		}
	}()

	for {
		conn, err := listener.Accept()
		if err != nil {
			return err
		}
		go s.handleConnection(conn)
	}
}

func (s *Server) Close() {
	s.stop <- struct{}{}
	<-s.stop
}

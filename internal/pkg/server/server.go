package server

import (
	"io"
	"net"

	"github.com/golang/glog"
	"github.com/inhzus/go-redis-impl/internal/pkg/command"
	"github.com/inhzus/go-redis-impl/internal/pkg/model"
	"github.com/inhzus/go-redis-impl/internal/pkg/token"
)

type Task struct {
	Cli *command.Client
	Req *token.Token
	Rsp chan *token.Token
}

type Option struct {
	Proto   string
	Addr    string
	DBCount int
}

type Server struct {
	option *Option
	queue  chan *Task
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
	if option.DBCount == 0 {
		option.DBCount = 16
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
	c := make(chan *token.Token)
	s.queue <- &Task{Req: t, Rsp: c}
	reply := <-c
	return reply
}

func (s *Server) handleConnection(conn net.Conn) {
	glog.Infof("client %v connection established", conn.RemoteAddr())
	cli := command.NewClient(conn)
	for {
		ts, err := token.Deserialize(conn)
		if err != nil {
			if err == io.EOF {
				glog.Infof("client %v connection closed", conn.RemoteAddr())
				return
			}
			glog.Error(err)
		}
		var data []byte
		for _, req := range ts {
			glog.Infof("request: %v", req.Format())
			c := make(chan *token.Token)
			s.queue <- &Task{Cli: cli, Req: req, Rsp: c}
			reply := <-c
			rsp, err := reply.Serialize()
			if err != nil {
				glog.Error(err)
				rsp, _ = token.ErrorDefault.Serialize()
			}
			data = append(data, rsp...)
		}
		go func(r []byte) {
			_, err = conn.Write(r)
			if err != nil {
				glog.Error(err)
			}
		}(data)
	}
}

func (s *Server) Serve() error {
	model.Init(s.option.DBCount)
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
				t.Rsp <- command.Process(t.Cli, t.Req)
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

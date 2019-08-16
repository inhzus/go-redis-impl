package server

import (
	"io"
	"net"

	"github.com/golang/glog"
	"github.com/inhzus/go-redis-impl/internal/pkg/command"
	"github.com/inhzus/go-redis-impl/internal/pkg/model"
	"github.com/inhzus/go-redis-impl/internal/pkg/token"
)

// Task is required to connect between handler & consumer
type Task interface {
	Task() Task
}

type InternalTask struct {
	Cmd     string
	DataIdx int
}

type CommandTask struct {
	Cli *model.Client
	Req *token.Token
	Rsp chan *token.Token
}

func (t *InternalTask) Task() Task { return t }
func (t *CommandTask) Task() Task  { return t }

// Option stores server configuration
type Option struct {
	Proto   string
	Addr    string
	DBCount int
}

// Server stores option, task queue & stop signal
type Server struct {
	option *Option
	queue  chan Task
	stop   chan struct{}
	proc   *command.Processor
}

// NewServer returns a new server pointer with default config
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

func (s *Server) handleConnection(conn net.Conn) {
	glog.Infof("client %v connection established", conn.RemoteAddr())
	cli := model.NewClient(conn, s.proc.Data[0])
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
			s.queue <- &CommandTask{Cli: cli, Req: req, Rsp: c}
			reply := <-c
			rsp, err := reply.Serialize()
			if err != nil {
				glog.Error(err)
				rsp, _ = token.NewError(err.Error()).Serialize()
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

// Serve starts handle connections synchronously
func (s *Server) Serve() error {
	s.proc = command.NewProcessor(s.option.DBCount)
	listener, err := net.Listen(s.option.Proto, s.option.Addr)
	if err != nil {
		return err
	}
	s.queue = make(chan Task)
	s.stop = make(chan struct{})

	go func() {
		for {
			select {
			case <-s.stop:
				_ = listener.Close()
				glog.Infof("server closed")
				s.stop <- struct{}{}
				return
			case task := <-s.queue:
				switch t := task.(type) {
				case *CommandTask:
					t.Rsp <- s.proc.Exec(t.Cli, t.Req)
				case *InternalTask:
					glog.Infof("command: %s on data_%d", t.Cmd, t.DataIdx)
				}
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

// Close stops server
func (s *Server) Close() {
	s.stop <- struct{}{}
	<-s.stop
}

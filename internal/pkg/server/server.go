package server

import (
	"io"
	"net"
	"time"

	"github.com/golang/glog"
	"github.com/inhzus/go-redis-impl/internal/pkg/model"
	"github.com/inhzus/go-redis-impl/internal/pkg/proc"
	"github.com/inhzus/go-redis-impl/internal/pkg/task"
	"github.com/inhzus/go-redis-impl/internal/pkg/token"
)

// Option stores server configuration
type Option struct {
	Addr    string
	DBCount int
	Persist struct {
		AppendName string
		CloneName  string
		FlushInr   time.Duration
		RewriteInr time.Duration
	}
	Proto string
}

// Server stores option, task queue & stop signal
type Server struct {
	option *Option
	queue  chan task.Task
	stop   chan struct{}
	proc   *proc.Processor
	setCh  chan *model.SetMsg
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
	if len(option.Persist.AppendName) == 0 {
		option.Persist.AppendName = "append-only.aof"
	}
	if len(option.Persist.CloneName) == 0 {
		option.Persist.CloneName = "data.rcl"
	}
	if option.Persist.FlushInr == 0 {
		option.Persist.FlushInr = time.Second
	}
	if option.Persist.RewriteInr == 0 {
		option.Persist.RewriteInr = time.Hour
	}
	return &Server{option: option}
}

func (s *Server) handleConnection(conn net.Conn) {
	glog.Infof("client %v connection established", conn.RemoteAddr())
	cli := s.proc.NewClient(conn, 0, s.setCh)
	for {
		ts, err := token.Deserialize(conn)
		var data []byte
		if err != nil {
			if err == io.EOF {
				glog.Infof("client %v connection closed", conn.RemoteAddr())
				return
			}
			glog.Error(err)
			data, _ = token.NewError(err.Error()).Serialize()
		}
		for _, req := range ts {
			glog.Infof("request: %v", req.Format())
			c := make(chan *token.Token)
			s.queue <- &model.CmdTask{Cli: cli, Req: req, Rsp: c}
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
func (s *Server) Serve() {
	listener, err := net.Listen(s.option.Proto, s.option.Addr)
	checkErr(err)
	s.queue = make(chan task.Task)
	s.stop = make(chan struct{})
	s.setCh = make(chan *model.SetMsg, 2<<10)
	s.proc = proc.NewProcessor(s.option.DBCount)
	go func() {
		for {
			select {
			case <-s.stop:
				_ = listener.Close()
				glog.Infof("server closed")
				s.stop <- struct{}{}
				return
			case tsk := <-s.queue:
				s.proc.Do(tsk)
			}
		}
	}()
	s.restoreData()
	go s.persistence()
	for {
		conn, err := listener.Accept()
		if err != nil {
			break
		}
		go s.handleConnection(conn)
	}
}

// Close stops server
func (s *Server) Close() {
	s.stop <- struct{}{}
	<-s.stop
}

package client

import (
	"net"
	"time"

	"github.com/golang/glog"
	"github.com/inhzus/go-redis-impl/internal/pkg/command"
	"github.com/inhzus/go-redis-impl/internal/pkg/token"
)

type Option struct {
	Addr  string
	Proto string
}

type Client struct {
	option *Option
	conn   net.Conn
	queue  chan *token.Task
	stop   chan struct{}
}

func NewClient(option *Option) *Client {
	if option.Proto == "" {
		option.Proto = "tcp"
	}
	if option.Proto == "unix" && option.Addr == "" {
		option.Addr = "/tmp/redis.sock"
	} else if option.Addr == "" {
		option.Addr = ":6389"
	}
	return &Client{option: option}
}

func (c *Client) process(t *token.Token) *token.Token {
	data, err := t.Serialize()
	if err != nil {
		return token.NewError("client cannot serialize token")
	}
	_, err = c.conn.Write(data)
	if err != nil {
		return token.NewError("connection cannot be write")
	}
	rsp, _ := token.Deserialize(c.conn)
	return rsp
}

func (c *Client) Connect() (err error) {
	c.conn, err = net.Dial(c.option.Proto, c.option.Addr)
	if err != nil {
		return
	}
	c.queue = make(chan *token.Task)
	c.stop = make(chan struct{})

	go func() {
		for {
			select {
			case <-c.stop:
				c.stop <- struct{}{}
				return
			case t := <-c.queue:
				t.Rsp <- c.process(t.Req)
			case <-time.After(time.Second * 1):
				c.process(token.NewString(command.CmdPing))
			}
		}
	}()
	return
}

func (c *Client) Close() {
	if c.conn != nil {
		_ = c.conn.Close()
		c.stop <- struct{}{}
		<-c.stop
	}
}

func (c *Client) Request(t *token.Token) *token.Token {
	if c.conn == nil {
		if err := c.Connect(); err != nil {
			return nil
		}
		defer c.Close()
	}
	ch := make(chan *token.Token)
	c.queue <- &token.Task{Req: t, Rsp: ch}
	rsp := <-ch
	glog.Infof("response: %v", rsp.Format())
	return rsp
}

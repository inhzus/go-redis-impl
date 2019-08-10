package client

import (
	"fmt"
	"net"
	"time"

	"github.com/golang/glog"
	"github.com/inhzus/go-redis-impl/internal/pkg/token"
)

type Option struct {
	Addr         string
	Proto        string
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
}

type Client struct {
	option  *Option
	conn    net.Conn
	queue   chan *token.Task
	stop    chan struct{}
	request func(*token.Token) *token.Response
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
	c := &Client{option: option}
	c.request = c.req
	return c
}

func (c *Client) consume(t *token.Token) *token.Response {
	data, err := t.Serialize()
	if err != nil {
		return &token.Response{Err: err}
	}
	if c.option.WriteTimeout > 0 {
		errCh := make(chan error)
		go func() {
			_, er := c.conn.Write(data)
			errCh <- er
		}()
		select {
		case <-time.After(c.option.WriteTimeout):
			return &token.Response{Err: fmt.Errorf("write data to connection timeout")}
		case err = <-errCh:
		}
	} else {
		_, err = c.conn.Write(data)
	}
	if err != nil {
		return &token.Response{Err: err}
	}
	var rsp *token.Response
	if c.option.ReadTimeout > 0 {
		rspCh := make(chan *token.Response)
		go func() {
			r := token.Deserialize(c.conn)
			rspCh <- r
		}()
		select {
		case <-time.After(c.option.ReadTimeout):
			return &token.Response{Err: fmt.Errorf("read data from connection timeout")}
		case rsp = <-rspCh:
		}
	} else {
		rsp = token.Deserialize(c.conn)
	}
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
				t.Rsp <- c.consume(t.Req)
			//case <-time.After(time.Second * 15):
			//	_, err = c.consume(token.NewArray(token.NewString(command.CmdPing)))
			//	if err != nil {
			//		glog.Errorf("failed to keep connection: %v", err)
			//	}
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

func (c *Client) Submit(t *token.Token) *token.Response {
	if c.conn == nil {
		defer c.Close()
		if err := c.Connect(); err != nil {
			return &token.Response{Err: err}
		}
	}
	ch := make(chan *token.Response)
	c.queue <- &token.Task{Req: t, Rsp: ch}
	rsp := <-ch
	if rsp.Err == nil {
		glog.Infof("%v-> %v", t.Format(), rsp.Data.Format())
	} else {
		glog.Errorf("%v-> %v", t.Format(), rsp.Err.Error())
	}
	return rsp
}

func (c *Client) req(t *token.Token) *token.Response {
	return c.Submit(t)
}

func (c *Client) Pipeline() *Pipeline {
	p := &Pipeline{Client: c}
	p.request = p.req
	p.Client.request = p.req
	return p
}

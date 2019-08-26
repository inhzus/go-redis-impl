package client

import (
	"bytes"
	"fmt"
	"net"
	"time"

	"github.com/inhzus/go-redis-impl/internal/pkg/cds"
	"github.com/inhzus/go-redis-impl/internal/pkg/token"
)

type Response struct {
	Data *token.Token
	Err  error
}

type Task struct {
	Req []*token.Token
	Rsp chan *Response
}

type Option struct {
	Addr         string
	Database     int
	Proto        string
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
}

type Client struct {
	option  *Option
	Conn    net.Conn
	queue   chan *Task
	stop    chan struct{}
	request func(*token.Token) *Response
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

func (c *Client) consume(ts []*token.Token, rspCh chan *Response) {
	defer close(rspCh)
	buffer := &bytes.Buffer{}
	for _, t := range ts {
		d, err := t.Serialize()
		if err != nil {
			rspCh <- &Response{Err: err}
			return
		}
		buffer.Write(d)
	}
	data := buffer.Bytes()
	var err error
	if c.option.WriteTimeout > 0 {
		_ = c.Conn.SetWriteDeadline(time.Now().Add(c.option.WriteTimeout))
	}
	_, err = c.Conn.Write(data)
	if err != nil {
		rspCh <- &Response{Err: err}
		return
	}
	endCh := make(chan struct{})
	var responses []*Response
	go func() {
		rsp, _ := token.Deserialize(c.Conn)
		for _, t := range rsp {
			responses = append(responses, &Response{Data: t})
		}
		endCh <- struct{}{}
	}()
	if c.option.ReadTimeout > 0 {
		select {
		case <-endCh:
		case <-time.After(c.option.ReadTimeout):
			rspCh <- &Response{Err: fmt.Errorf("read data from connection timeout")}
			return
		}
	} else {
		<-endCh
	}
	for _, rsp := range responses {
		rspCh <- rsp
	}
	return
}

func (c *Client) Connect() (err error) {
	c.Conn, err = net.DialTimeout(c.option.Proto, c.option.Addr, time.Second)
	if err != nil {
		return
	}
	c.queue = make(chan *Task)
	c.stop = make(chan struct{})

	go func() {
		for {
			select {
			case <-c.stop:
				c.stop <- struct{}{}
				return
			case t := <-c.queue:
				c.consume(t.Req, t.Rsp)
			}
		}
	}()
	if c.option.Database > 0 {
		c.Submit(token.NewArray(token.NewString(cds.Select), token.NewInteger(int64(c.option.Database))))
	}
	return
}

func (c *Client) Close() {
	if c.Conn != nil {
		_ = c.Conn.Close()
		c.stop <- struct{}{}
		<-c.stop
	}
}

func (c *Client) Submit(t ...*token.Token) <-chan *Response {
	ch := make(chan *Response, 1)
	if c.Conn == nil {
		go func() { ch <- &Response{Err: fmt.Errorf("connection is nil")} }()
		return ch
	}
	c.queue <- &Task{Req: t, Rsp: ch}
	return ch
}

func (c *Client) req(t *token.Token) *Response {
	rsp := <-c.Submit(t)
	return rsp
}

func (c *Client) Pipeline() *Pipeline {
	p := &Pipeline{Client: *c}
	p.request = p.req
	p.Client.request = p.req
	return p
}

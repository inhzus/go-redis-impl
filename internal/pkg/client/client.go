package client

import (
	"net"

	"github.com/golang/glog"
	"github.com/inhzus/go-redis-impl/internal/pkg/token"
)

type Option struct {
	Addr  string
	Proto string
}

type Client struct {
	option  *Option
	conn    net.Conn
	queue   chan *token.Task
	stop    chan struct{}
	request func(*token.Token) (*token.Token, error)
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

func (c *Client) consume(t *token.Token) (*token.Token, error) {
	data, err := t.Serialize()
	if err != nil {
		return nil, err
	}
	_, err = c.conn.Write(data)
	if err != nil {
		return nil, err
	}
	rsp, _ := token.Deserialize(c.conn)
	glog.Infof("response: %v", rsp.Format())
	return rsp, err
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
				d, err := c.consume(t.Req)
				t.Rsp <- &token.Response{Data: d, Err: err}
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

func (c *Client) Submit(t *token.Token) (*token.Token, error) {
	if c.conn == nil {
		defer c.Close()
		if err := c.Connect(); err != nil {
			return nil, err
		}
	}
	ch := make(chan *token.Response)
	c.queue <- &token.Task{Req: t, Rsp: ch}
	rsp := <-ch
	return rsp.Data, rsp.Err
}

func (c *Client) req(t *token.Token) (*token.Token, error) {
	return c.Submit(t)
}

func (c *Client) Pipeline() *Pipeline {
	p := &Pipeline{Client: c}
	p.request = p.req
	p.Client.request = p.req
	return p
}

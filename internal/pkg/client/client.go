package client

import (
	"bufio"
	"net"

	"github.com/golang/glog"
	"github.com/inhzus/go-redis-impl/internal/pkg/token"
)

type Option struct {
	Addr  string
	Proto string
}

type Client struct {
	option *Option
	conn   net.Conn
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
	return &Client{option: option,}
}

func (c *Client) Connect() (err error) {
	c.conn, err = net.Dial(c.option.Proto, c.option.Addr)
	if err != nil {
		return
	}
	return
}

func (c *Client) Close() error {
	return c.conn.Close()
}

func (c *Client) Request(argument *token.Argument) {
	if c.conn == nil {
		if err := c.Connect(); err != nil {
			glog.Error(err)
			return
		}
		defer func() { _ = c.Close() }()
	}
	data, err := token.Compose(argument)
	if err != nil {
		glog.Error(err)
		return
	}
	_, _ = c.conn.Write(data)
	rsp, _ := bufio.NewReader(c.conn).ReadBytes('\n')
	glog.Infof("response: %v", string(rsp))
}

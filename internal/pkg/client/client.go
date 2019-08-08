package client

import (
	"net"

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
	return &Client{option: option}
}

func (c *Client) Connect() (err error) {
	c.conn, err = net.Dial(c.option.Proto, c.option.Addr)
	if err != nil {
		return
	}
	return
}

func (c *Client) Close() error {
	if c.conn != nil {
		return c.conn.Close()
	} else {
		return nil
	}
}

func (c *Client) Request(t *token.Token) (*token.Token, error) {
	if c.conn == nil {
		if err := c.Connect(); err != nil {
			return nil, err
		}
		defer func() { _ = c.Close() }()
	}
	data, err := t.Serialize()
	if err != nil {
		return nil, err
	}
	_, err = c.conn.Write(data)
	if err != nil {
		return nil, err
	}
	rsp, _ := token.Deserialize(c.conn)
	return rsp, nil
}

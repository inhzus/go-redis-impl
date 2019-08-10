package client

import (
	"github.com/inhzus/go-redis-impl/internal/pkg/command"
	"github.com/inhzus/go-redis-impl/internal/pkg/token"
)

func (c *Client) Get(key string) *token.Response {
	row := token.NewArray(token.NewString(command.CmdGet), token.NewString(key))
	return c.request(row)
}

func (c *Client) Set(key string, value interface{}) *token.Response {
	bulked, err := command.ItfToBulked(value)
	if err != nil {
		return &token.Response{Err: err}
	}
	row := token.NewArray(token.NewString(command.CmdSet), token.NewString(key), token.NewBulked(bulked))
	return c.request(row)
}

func (c *Client) Ping() *token.Response {
	row := token.NewArray(token.NewString(command.CmdPing))
	return c.request(row)
}

func (c *Client) Incr(key string) *token.Response {
	row := token.NewArray(token.NewString(command.CmdIncr), token.NewString(key))
	return c.request(row)
}

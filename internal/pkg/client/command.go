package client

import (
	"time"

	"github.com/inhzus/go-redis-impl/internal/pkg/command"
	"github.com/inhzus/go-redis-impl/internal/pkg/token"
)

func (c *Client) Get(key string) *token.Response {
	row := token.NewArray(token.NewString(command.CmdGet), token.NewString(key))
	return c.request(row)
}

func (c *Client) Set(key string, value interface{}, timeout time.Duration) *token.Response {
	bulked, err := command.ItfToBulked(value)
	if err != nil {
		return &token.Response{Err: err}
	}
	row := token.NewArray(token.NewString(command.CmdSet), token.NewString(key), token.NewBulked(bulked))
	if timeout > 0 {
		row.Data = append(row.Data.([]*token.Token),
			token.NewString(command.TimeoutMilSec), token.NewInteger(int64(timeout/time.Millisecond)))
	}
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

func (c *Client) Desc(key string) *token.Response {
	row := token.NewArray(token.NewString(command.CmdDesc), token.NewString(key))
	return c.request(row)
}

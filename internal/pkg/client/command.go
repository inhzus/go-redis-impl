package client

import (
	"time"

	"github.com/inhzus/go-redis-impl/internal/pkg/proc"
	"github.com/inhzus/go-redis-impl/internal/pkg/token"
)

func (c *Client) Get(key string) *Response {
	row := token.NewArray(token.NewString(proc.CmdGet), token.NewString(key))
	return c.request(row)
}

func (c *Client) Set(key string, value interface{}, timeout time.Duration) *Response {
	bulked, err := proc.ItfToBulked(value)
	if err != nil {
		return &Response{Err: err}
	}
	row := token.NewArray(token.NewString(proc.CmdSet), token.NewString(key), token.NewBulked(bulked))
	if timeout > 0 {
		row.Data = append(row.Data.([]*token.Token),
			token.NewString(proc.TimeoutMilSec), token.NewInteger(int64(timeout/time.Millisecond)))
	}
	return c.request(row)
}

func (c *Client) Ping() *Response {
	row := token.NewArray(token.NewString(proc.CmdPing))
	return c.request(row)
}

func (c *Client) Incr(key string) *Response {
	row := token.NewArray(token.NewString(proc.CmdIncr), token.NewString(key))
	return c.request(row)
}

func (c *Client) Desc(key string) *Response {
	row := token.NewArray(token.NewString(proc.CmdDesc), token.NewString(key))
	return c.request(row)
}

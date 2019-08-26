package client

import (
	"time"

	"github.com/inhzus/go-redis-impl/internal/pkg/cds"
	"github.com/inhzus/go-redis-impl/internal/pkg/proc"
	"github.com/inhzus/go-redis-impl/internal/pkg/token"
)

// Redis `get` command.
func (c *Client) Get(key string) *Response {
	row := token.NewArray(token.NewString(cds.Get), token.NewString(key))
	return c.request(row)
}

// Redis `set` command.
func (c *Client) Set(key string, value interface{}, timeout time.Duration) *Response {
	bulked, err := proc.ItfToBulked(value)
	if err != nil {
		return &Response{Err: err}
	}
	row := token.NewArray(token.NewString(cds.Set), token.NewString(key), token.NewBulked(bulked))
	if timeout > 0 {
		row.Data = append(row.Data.([]*token.Token),
			token.NewString(cds.TimeoutMilSec), token.NewInteger(int64(timeout/time.Millisecond)))
	}
	return c.request(row)
}

// Redis `ping` command.
func (c *Client) Ping() *Response {
	row := token.NewArray(token.NewString(cds.Ping))
	return c.request(row)
}

// Redis `incr` command.
func (c *Client) Incr(key string) *Response {
	row := token.NewArray(token.NewString(cds.Incr), token.NewString(key))
	return c.request(row)
}

// Redis `desc` command.
func (c *Client) Desc(key string) *Response {
	row := token.NewArray(token.NewString(cds.Desc), token.NewString(key))
	return c.request(row)
}

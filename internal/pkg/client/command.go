package client

import (
	"github.com/inhzus/go-redis-impl/internal/pkg/command"
	"github.com/inhzus/go-redis-impl/internal/pkg/token"
)

func (c *Client) Get(key string) (*token.Token, error) {
	row := token.NewArray(token.NewString(command.CmdGet), token.NewString(key))
	return c.request(row)
}

func (c *Client) Set(key string, value []byte) (*token.Token, error) {
	row := token.NewArray(token.NewString(command.CmdSet), token.NewString(key), token.NewBulked(value))
	return c.request(row)
}

func (c *Client) Ping() (*token.Token, error) {
	row := token.NewArray(token.NewString(command.CmdPing))
	return c.request(row)
}

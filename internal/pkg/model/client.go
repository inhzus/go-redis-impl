package model

import (
	"net"

	"github.com/inhzus/go-redis-impl/internal/pkg/token"
)

// MultiInfo stores transaction info related
type MultiInfo struct {
	// true if transaction started
	State bool
	// true if watched key is changed
	Dirty bool
	// transaction queue
	Queue []*token.Token
}

// Client entity: client information stored
type Client struct {
	Conn net.Conn
	// database index selected
	DataIdx int
	// transaction info
	Multi *MultiInfo
}

// NewClient returns a client selecting database 0, transaction state false
func NewClient(conn net.Conn) *Client {
	return &Client{Conn: conn, Multi: &MultiInfo{}}
}

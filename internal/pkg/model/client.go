package model

import (
	"net"

	"github.com/inhzus/go-redis-impl/internal/pkg/token"
)

// Client entity: client information stored
type Client struct {
	Conn net.Conn
	// database index selected
	DataIdx int
	// transaction state, true if transaction started
	MultiState bool
	// transaction queue
	Queue []*token.Token
}

// NewClient returns a client selecting database 0, transaction state false
func NewClient(conn net.Conn) *Client {
	return &Client{Conn: conn, DataIdx: 0, MultiState: false}
}

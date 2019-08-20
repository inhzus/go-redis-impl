package model

import (
	"net"

	"github.com/inhzus/go-redis-impl/internal/pkg/token"
)

// MultiInfo stores transaction info related
type MultiInfo struct {
	// true if transaction started
	State bool
	// true if watched Key is changed
	Dirty bool
	// transaction queue
	Queue []*token.Token
	// watch keys
	Watched []*watchKey
}

// Client entity: client information stored
type Client struct {
	Conn net.Conn
	// database index selected
	Data *DataStorage
	// transaction info
	Multi *MultiInfo
}

// NewClient returns a client selecting database 0, transaction state false
func NewClient(conn net.Conn, dataStorage *DataStorage) *Client {
	return &Client{Conn: conn, Data: dataStorage, Multi: &MultiInfo{}}
}

// Watch append Key to self watch list and append self to global watch map
func (c *Client) Watch(key string) {
	for _, v := range c.Multi.Watched {
		if v.wc.key == key {
			return
		}
	}
	c.Multi.Watched = append(c.Multi.Watched, c.Data.watch.Put(c, key))
}

// Unwatch cancel all watched keys
func (c *Client) Unwatch() {
	for _, v := range c.Multi.Watched {
		c.Data.watch.Remove(v)
	}
	c.Multi.Watched = nil
	c.Multi.Dirty = false
}

// Get returns correspond value of data indexed and Key
func (c *Client) Get(key string) interface{} {
	return c.Data.Get(key)
}

// Set puts Key-value pair and its ttl in data
func (c *Client) Set(key string, value interface{}, expire int64) interface{} {
	c.Data.watch.Touch(key)
	return c.Data.Set(key, value, expire)
}

// Del deletes the value of correspond Key
func (c *Client) Del(key string) {
	c.Data.Del(key)
}

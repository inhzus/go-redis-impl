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
	// watch keys
	Watched []*watchKey
}

// Client entity: client information stored
type Client struct {
	Conn net.Conn
	// database index selected
	Data *DataStorage
	Idx  int
	// transaction info
	Multi *MultiInfo
	SetCh chan *SetMsg
}

// NewClient returns a client selecting database 0, transaction state false
func NewClient(conn net.Conn, dataStorage *DataStorage, idx int, setCh chan *SetMsg) *Client {
	return &Client{Conn: conn, Data: dataStorage, Idx: idx, Multi: &MultiInfo{}, SetCh: setCh}
}

// Watch append key to self watch list and append self to global watch map
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

// Get returns correspond value of data indexed and key
func (c *Client) Get(key string) interface{} {
	return c.Data.Get(key)
}

// Set puts key-value pair and its ttl in data
func (c *Client) Set(key string, value interface{}, expire int64) interface{} {
	c.Data.watch.Touch(key)
	if c.SetCh != nil {
		t := token.NewArray(token.NewString("set"), token.NewString(key), token.NewToken(value))
		if expire > 0 {
			t.Data = append(t.Data.([]*token.Token), token.NewString("PT"), token.NewInteger(expire))
		}
		c.SetCh <- &SetMsg{Idx: c.Idx, T: t}
	}
	return c.Data.Set(key, value, expire)
}

// Del deletes the value of correspond key
func (c *Client) Del(key string) {
	c.Data.Del(key)
}

package model

import (
	"time"
)

type Item struct {
	key    string
	row    interface{}
	ttl    time.Duration
	expire time.Time
	index  int
}

func newItem(key string, row interface{}, ttl time.Duration) *Item {
	item := &Item{key: key, row: row, ttl: ttl}
	if ttl > 0 {
		item.expire = time.Now().Add(ttl)
	}
	return item
}

type DataStorage struct {
	data map[string]*Item
}

func NewData() *DataStorage {
	return &DataStorage{make(map[string]*Item)}
}

func (d *DataStorage) Get(key string) interface{} {
	r, ok := d.data[key]
	if ok {
		return r.row
	} else {
		return nil
	}
}

func (d *DataStorage) Set(key string, value interface{}, ttl time.Duration) interface{} {
	item := newItem(key, value, ttl)
	d.data[key] = item
	return item.row
}

var (
	Data = NewData()
)

func Get(key string) interface{} {
	return Data.Get(key)
}

func Set(key string, value interface{}, ttl time.Duration) interface{} {
	return Data.Set(key, value, ttl)
}

package model

import (
	"container/heap"
	"container/list"
	"time"
)

const (
	checkExpireNum = 10
)

// Item is key-value pair stored in model
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

func (i *Item) fix(row interface{}, ttl time.Duration) {
	if ttl == 0 && i.expire.Before(time.Now()) {
		i.ttl = 0
	} else if ttl > 0 {
		i.ttl = ttl
		i.expire = time.Now().Add(ttl)
	}
	i.row = row
}

type dataStorage struct {
	data  map[string]*Item
	queue *priorityQueue
	watch map[string]list.List
}

func (d *dataStorage) scanPop(n int) {
	now := time.Now()
	for i := 0; i < checkExpireNum; i++ {
		top := d.queue.Top()
		if top == nil {
			return
		}
		if top.ttl > 0 && top.expire.Before(now) {
			heap.Pop(d.queue)
			delete(d.data, top.key)
		} else {
			return
		}
	}
}

func (d *dataStorage) Get(key string) interface{} {
	d.scanPop(checkExpireNum)
	r, ok := d.data[key]
	if !ok {
		return nil
	}
	if r.ttl > 0 && r.expire.Before(time.Now()) {
		return nil
	}
	return r.row
}

func (d *dataStorage) Set(key string, value interface{}, ttl time.Duration) interface{} {
	d.scanPop(checkExpireNum)
	item, ok := d.data[key]
	if ok {
		item.fix(value, ttl)
		if item.ttl > 0 {
			heap.Fix(d.queue, item.index)
		}
	} else {
		item = newItem(key, value, ttl)
		if item.ttl > 0 {
			heap.Push(d.queue, item)
		}
		d.data[key] = item
	}
	return item.row
}

var (
	data []*dataStorage
)

// Init initializes data
func Init(n int) {
	data = make([]*dataStorage, n)
	for i := 0; i < n; i++ {
		data[i] = &dataStorage{make(map[string]*Item), &priorityQueue{}, make(map[string]list.List)}
	}
}

// Get returns correspond value of data indexed and key
func Get(cli *Client, key string) interface{} {
	return data[cli.DataIdx].Get(key)
}

// Set puts key-value pair and its ttl in data
func Set(cli *Client, key string, value interface{}, ttl time.Duration) interface{} {
	return data[cli.DataIdx].Set(key, value, ttl)
}

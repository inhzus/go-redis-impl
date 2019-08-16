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
	expire int64
	index  int
	elem   *list.Element
}

func newItem(key string, row interface{}, ttl time.Duration) *Item {
	item := &Item{key: key, row: row}
	if ttl > 0 {
		item.expire = time.Now().Add(ttl).UnixNano()
	}
	return item
}

func (i *Item) fix(row interface{}, ttl time.Duration) {
	if ttl == 0 && i.expire < time.Now().UnixNano() {
		i.expire = 0
	} else if ttl > 0 {
		i.expire = time.Now().Add(ttl).UnixNano()
	}
	i.row = row
}

// DataStorage stores key-value data, expiration control heap, watched key-client map
type DataStorage struct {
	data  map[string]*Item
	queue *priorityQueue
	watch watchMap
}

func NewDataStorage() *DataStorage {
	return &DataStorage{make(map[string]*Item), &priorityQueue{}, newWatchMap()}
}

func (d *DataStorage) scanPop(n int) {
	now := time.Now().UnixNano()
	for i := 0; i < checkExpireNum; i++ {
		top := d.queue.Top()
		if top == nil {
			return
		}
		if top.expire > 0 && top.expire < now {
			heap.Pop(d.queue)
			delete(d.data, top.key)
		} else {
			return
		}
	}
}

func (d *DataStorage) Get(key string) interface{} {
	d.scanPop(checkExpireNum)
	r, ok := d.data[key]
	if !ok {
		return nil
	}
	if r.expire > 0 && r.expire < time.Now().UnixNano() {
		return nil
	}
	return r.row
}

func (d *DataStorage) Set(key string, value interface{}, ttl time.Duration) interface{} {
	d.scanPop(checkExpireNum)
	item, ok := d.data[key]
	if ok {
		item.fix(value, ttl)
		if ttl > 0 {
			heap.Fix(d.queue, item.index)
		}
	} else {
		item = newItem(key, value, ttl)
		heap.Push(d.queue, item)
		d.data[key] = item
	}
	return item.row
}

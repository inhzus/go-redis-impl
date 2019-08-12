package model

import (
	"container/heap"
	"time"
)

const (
	CheckExpireNum = 10
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

func (i *Item) Fix(row interface{}, ttl time.Duration) {
	if ttl == 0 && i.expire.Before(time.Now()) {
		i.ttl = 0
	} else if ttl > 0 {
		i.ttl = ttl
		i.expire = time.Now().Add(ttl)
	}
	i.row = row
}

type DataStorage struct {
	data  map[string]*Item
	queue *priorityQueue
}

func (d *DataStorage) scanPop(n int) {
	now := time.Now()
	for i := 0; i < CheckExpireNum; i++ {
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

func (d *DataStorage) Get(key string) interface{} {
	d.scanPop(CheckExpireNum)
	r, ok := d.data[key]
	if !ok {
		return nil
	}
	if r.ttl > 0 && r.expire.Before(time.Now()) {
		return nil
	}
	return r.row
}

func (d *DataStorage) Set(key string, value interface{}, ttl time.Duration) interface{} {
	d.scanPop(CheckExpireNum)
	item, ok := d.data[key]
	if ok {
		item.Fix(value, ttl)
		heap.Fix(d.queue, item.index)
	} else {
		item = newItem(key, value, ttl)
		heap.Push(d.queue, item)
		d.data[key] = item
	}
	return item.row
}

var (
	Data []*DataStorage
)

func Init(n int) {
	Data = make([]*DataStorage, n)
	for i := 0; i < n; i++ {
		Data[i] = &DataStorage{make(map[string]*Item), &priorityQueue{}}
	}
}

func Get(idx int, key string) interface{} {
	return Data[idx].Get(key)
}

func Set(idx int, key string, value interface{}, ttl time.Duration) interface{} {
	return Data[idx].Set(key, value, ttl)
}

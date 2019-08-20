package model

import (
	"container/heap"
	"container/list"
	"time"
)

const (
	checkExpireNum = 10
	moveBackNum    = 10
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

func newExpiredItem(key string) *Item {
	return &Item{key: key, expire: time.Now().UnixNano()}
}

func (i *Item) fix(row interface{}, ttl time.Duration) {
	if ttl == 0 && i.expire < time.Now().UnixNano() {
		i.expire = 0
	} else if ttl > 0 {
		i.expire = time.Now().Add(ttl).UnixNano()
	}
	i.row = row
}

func (i *Item) makeExpired() {
	i.expire = time.Now().UnixNano()
}

// DataStorage stores key-value data, expiration control heap, watched key-client map
type DataStorage struct {
	data     map[string]*Item
	oldData  map[string]*Item
	queue    *priorityQueue
	oldQueue *priorityQueue
	watch    watchMap
	isMoving bool
	isBlock  bool
}

// NewDataStorage returns data storage entity with default constructor
func NewDataStorage() *DataStorage {
	return &DataStorage{data: make(map[string]*Item), queue: &priorityQueue{}, watch: newWatchMap()}
}

func (d *DataStorage) freeze() {
	d.isBlock = true
	d.oldData = d.data
	d.data = make(map[string]*Item)
	d.oldQueue = d.queue
	d.queue = &priorityQueue{}
}

func (d *DataStorage) toMove() {
	d.isBlock = false
	d.isMoving = true
}

func (d *DataStorage) scanPop(n int) {
	if d.isBlock {
		return
	}
	queue := d.queue
	data := d.data
	if d.isMoving {
		queue = d.oldQueue
		data = d.oldData
	}
	now := time.Now().UnixNano()
	for i := 0; i < checkExpireNum; i++ {
		top := queue.Top()
		if top == nil {
			return
		}
		if top.expire > 0 && top.expire < now {
			heap.Pop(queue)
			delete(data, top.key)
		} else {
			return
		}
	}
}

func (d *DataStorage) moveBack(n int) {
	for i := 0; i < n && d.queue.Len() > 0; i++ {
		top := d.queue.Top()
		heap.Pop(d.queue)
		delete(d.data, top.key)
		d.oldData[top.key] = top
		heap.Push(d.oldQueue, top)
	}
}

func (d *DataStorage) resetIfMoved() bool {
	if !d.isMoving {
		return true
	}
	if len(d.data) != 0 {
		return false
	}
	d.isMoving = false
	d.data, d.oldData = d.oldData, nil
	d.queue, d.oldQueue = d.oldQueue, nil
	return true
}

func (d *DataStorage) Get(key string) interface{} {
	d.scanPop(checkExpireNum)
	r, ok := d.data[key]
	if !ok && (d.isBlock || d.isMoving) {
		r, ok = d.oldData[key]
	}
	if !ok {
		return nil
	}
	now := time.Now().UnixNano()
	if r.expire > 0 && r.expire < now {
		return nil
	}
	return r.row
}

func (d *DataStorage) Set(key string, value interface{}, ttl time.Duration) interface{} {
	d.scanPop(checkExpireNum)
	data := d.data
	queue := d.queue
	if !d.resetIfMoved() {
		d.moveBack(moveBackNum)
		item, ok := d.data[key]
		if ok {
			heap.Remove(d.queue, item.index)
			delete(d.data, key)
		}
		data = d.oldData
		queue = d.oldQueue
	}

	item, ok := data[key]
	if ok {
		item.fix(value, ttl)
		heap.Fix(queue, item.index)
	} else {
		item = newItem(key, value, ttl)
		data[key] = item
		heap.Push(queue, item)
	}
	return item.row
}

func (d *DataStorage) Del(key string) {
	item, ok := d.data[key]
	if d.isBlock {
		if ok {
			item.makeExpired()
			heap.Fix(d.queue, item.index)
		} else {
			item = newExpiredItem(key)
			d.data[key] = item
			heap.Push(d.queue, item)
		}
		return
	}
	if ok {
		heap.Remove(d.queue, item.index)
		delete(d.data, key)
	}
	if d.isMoving {
		item, ok = d.oldData[key]
		if ok {
			heap.Remove(d.oldQueue, item.index)
			delete(d.oldData, key)
		}
	}
}

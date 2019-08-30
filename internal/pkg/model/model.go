package model

import (
	"container/heap"
	"fmt"
	"time"
)

const (
	checkExpireNum = 10
	moveBackNum    = 10
)

// Item is key-value pair stored in model
type Item struct {
	key    string
	Row    interface{}
	Expire int64
	index  int
}

func newItem(key string, row interface{}, expire int64) *Item {
	return &Item{key: key, Row: row, Expire: expire}
}

// newExpiredItem returns a new item that is expired.
// The function is used when origin data is blocked.
// The item taking the place ensures that origin item is set expired after moving.
func newExpiredItem(key string) *Item {
	return &Item{key: key, Expire: time.Now().UnixNano() - 1}
}

func (i *Item) fix(row interface{}, expire int64) {
	if expire == 0 && i.Expire < time.Now().UnixNano() {
		i.Expire = 0
	} else if expire > 0 {
		i.Expire = expire
	}
	i.Row = row
}

func (i *Item) makeExpired() {
	i.Expire = time.Now().UnixNano() - 1
}

// DataStorage stores key-value data, expiration control heap, watched key-client map
type DataStorage struct {
	data     map[string]*Item
	oldData  map[string]*Item
	queue    *priorityQueue
	oldQueue *priorityQueue
	watch    watchMap
	// blocked data has two status, status block means origin data should not change,
	// status moving means moving new data back to origin data
	isMoving bool
	isBlock  bool
}

// NewDataStorage returns data storage entity with default constructor
func NewDataStorage() *DataStorage {
	return &DataStorage{data: make(map[string]*Item), queue: &priorityQueue{}, watch: newWatchMap()}
}

// GetOrigin returns the original data of the data storage.
func (d *DataStorage) GetOrigin() map[string]*Item {
	return d.oldData
}

// Freeze and following logic ensure the origin data won't change until "ToMove"
func (d *DataStorage) Freeze() error {
	if !d.resetIfMoved() {
		return fmt.Errorf("data is moving")
	}
	d.isBlock = true
	d.oldData = d.data
	d.data = make(map[string]*Item)
	d.oldQueue = d.queue
	d.queue = &priorityQueue{}
	return nil
}

// ToMove notifies data storage to move items from new data to old data
func (d *DataStorage) ToMove() error {
	if !d.isBlock {
		return fmt.Errorf("data was not blocked")
	}
	d.isBlock = false
	d.isMoving = true
	return nil
}

// scanPop checks and pops n * expired items when called.
func (d *DataStorage) scanPop(n int) {
	// When origin data blocked, expired item should be kept to set origin item expired when moving.
	if d.isBlock {
		return
	}
	queue := &d.queue
	data := &d.data
	// When moving, origin data is able to check expired again.
	if d.isMoving {
		queue = &d.oldQueue
		data = &d.oldData
	}
	now := time.Now().UnixNano()
	for i := 0; i < checkExpireNum; i++ {
		top := (*queue).Top()
		if top == nil {
			return
		}
		if top.Expire > 0 && top.Expire < now {
			heap.Pop(*queue)
			delete(*data, top.key)
		} else {
			return
		}
	}
}

// moveBack moves n item from new data to origin data when called
func (d *DataStorage) moveBack(n int) {
	for i := 0; i < n && d.queue.Len() > 0; i++ {
		top := d.queue.Top()
		heap.Pop(d.queue)
		delete(d.data, top.key)
		d.oldData[top.key] = top
		heap.Push(d.oldQueue, top)
	}
}

// resetIfMoved checks whether status moving is finished and resets the data
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

// Get returns the value of correspond key
func (d *DataStorage) Get(key string) interface{} {
	d.scanPop(checkExpireNum)
	if !d.resetIfMoved() {
		d.moveBack(moveBackNum)
	}
	r, ok := d.data[key]
	// when blocked or moving, both data should be checked
	if !ok && (d.isBlock || d.isMoving) {
		r, ok = d.oldData[key]
	}
	if !ok {
		return nil
	}
	now := time.Now().UnixNano()
	if r.Expire > 0 && r.Expire < now {
		return nil
	}
	return r.Row
}

// Set puts the new value of key
func (d *DataStorage) Set(key string, value interface{}, expire int64) interface{} {
	d.scanPop(checkExpireNum)
	data := &d.data
	queue := &d.queue
	if !d.resetIfMoved() {
		d.moveBack(moveBackNum)
		// When moving, set data into origin data and delete the one in new data.
		item, ok := d.data[key]
		if ok {
			heap.Remove(d.queue, item.index)
			delete(d.data, key)
		}
		data = &d.oldData
		queue = &d.oldQueue
	}

	item, ok := (*data)[key]
	if ok {
		item.fix(value, expire)
		heap.Fix(*queue, item.index)
	} else {
		item = newItem(key, value, expire)
		(*data)[key] = item
		heap.Push(*queue, item)
	}
	return item.Row
}

// Del deletes the value of correspond key
func (d *DataStorage) Del(key string) {
	item, ok := d.data[key]
	// When blocked, origin data shouldn't be changed, just set item of correspond key in new data expired
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
	// When moving, both data in origin data & new data should be deleted
	if d.isMoving {
		item, ok = d.oldData[key]
		if ok {
			heap.Remove(d.oldQueue, item.index)
			delete(d.oldData, key)
		}
	}
}

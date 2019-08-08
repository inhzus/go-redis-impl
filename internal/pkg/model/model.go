package model

import "time"

type Item struct {
	row    interface{}
	expire time.Time
}

type DataStorage struct {
	data map[string]Item
}

func NewData() *DataStorage {
	return &DataStorage{make(map[string]Item)}
}

func (d *DataStorage) Get(key string) interface{} {
	return d.data[key].row
}

func (d *DataStorage) Set(key string, value interface{}, expire time.Time) interface{} {
	item := Item{row: value, expire: expire,}
	d.data[key] = item
	return item.row
}

var (
	Data = NewData()
)

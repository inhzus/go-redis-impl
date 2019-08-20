package model

import (
	"fmt"
	"net"
	"reflect"
	"strconv"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

var data = NewDataStorage()
var cli = NewClient(nil, data)

func TestNewClient(t *testing.T) {
	d := NewDataStorage()
	type args struct {
		conn net.Conn
		data *DataStorage
	}
	tests := []struct {
		name string
		args args
		want *Client
	}{
		{"new client", args{conn: nil, data: d}, &Client{Conn: nil, Data: d, Multi: &MultiInfo{false, false, nil, nil}}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewClient(tt.args.conn, tt.args.data); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewClient() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestClient_Watch(t *testing.T) {
	const Num = 100
	cs := make([]*Client, Num)
	for i := range cs {
		cs[i] = NewClient(nil, data)
	}
	for i := 0; i < 100; i++ {
		cs[i].Watch(strconv.FormatInt(int64(i%5), 10))
		cs[i].Watch(strconv.FormatInt(int64(i%10), 10))
		if i%10 < 5 {
			cs[i].Multi.State = true
		}
	}
	assert.Equal(t, 10, len(data.watch))
	for i := 0; i < 5; i++ {
		k := strconv.FormatInt(int64(i), 10)
		assert.Equal(t, 20, data.watch[k].clients.Len())
		assert.Equal(t, k, data.watch[k].key)
	}
	for i := 5; i < 10; i++ {
		k := strconv.FormatInt(int64(i), 10)
		assert.Equal(t, 10, data.watch[k].clients.Len())
		assert.Equal(t, k, data.watch[k].key)
	}
	for i := 50; i < 100; i++ {
		cs[i].Unwatch()
	}
	cli.Set("10", 1, 0)
	cli.Set("1", 1, 0)
	for i := 0; i < 100; i++ {
		if i%5 == 1 && i%10 < 5 && i < 50 {
			assert.True(t, cs[i].Multi.Dirty, fmt.Sprintf("index: %v", i))
		} else {
			assert.False(t, cs[i].Multi.Dirty, fmt.Sprintf("index: %v", i))
		}
	}
	for i := 0; i < 100; i++ {
		cs[i].Unwatch()
	}
	for i := 0; i < 100; i++ {
		assert.Nil(t, cs[i].Multi.Watched)
	}
	assert.Zero(t, len(data.watch))
	cli.Set("10", 1, time.Nanosecond)
	cli.Set("1", 1, time.Nanosecond)
	<-time.After(time.Millisecond)
}

func TestClient_Set(t *testing.T) {
	KeyFmt := "a_%d"
	for i := 0; i < 20; i++ {
		cli.Set(fmt.Sprintf(KeyFmt, i), i, time.Second)
	}
	for i := 0; i < 20; i++ {
		val := cli.Get(fmt.Sprintf(KeyFmt, i))
		assert.Equal(t, val.(int), i)
	}
	for i := 0; i < 10; i++ {
		cli.Set(fmt.Sprintf(KeyFmt, i), i+1, time.Second/2)
	}
	for i := 0; i < 10; i++ {
		val := cli.Get(fmt.Sprintf(KeyFmt, i))
		assert.Equal(t, val.(int), i+1)
	}
	<-time.After(time.Second / 2)
	for i := 0; i < 10; i++ {
		val := cli.Get(fmt.Sprintf(KeyFmt, i))
		assert.Equal(t, val, nil)
	}
	<-time.After(time.Second / 2)
	for i := 11; i < 20; i++ {
		val := cli.Get(fmt.Sprintf(KeyFmt, i))
		assert.Equal(t, val, nil)
	}
}

func TestClient(t *testing.T) {
	data.freeze()
	TestClient_Set(t)
	assert.Equal(t, 20, len(data.queue.items))
	assert.Equal(t, 0, len(data.oldQueue.items))
	assert.Equal(t, 20, len(data.data))
	assert.Equal(t, 0, len(data.oldData))
	TestClient_Get(t)
	assert.Equal(t, 20, len(data.queue.items))
	assert.Equal(t, 0, len(data.oldQueue.items))
	assert.Equal(t, 20, len(data.data))
	assert.Equal(t, 0, len(data.oldData))
	data.toMove()
	TestClient_Get(t)
	TestClient_Set(t)
}

func TestClient_Get(t *testing.T) {
	assert.Equal(t, len(data.data), data.queue.Len())
	KeyFmt := "a_%d"
	for i := 0; i < 20; i++ {
		cli.Set(fmt.Sprintf(KeyFmt, i), i, 0)
	}
	assert.Equal(t, cli.Get("a_15"), 15)
	for i := 0; i < 20; i++ {
		cli.Set(fmt.Sprintf(KeyFmt, i), i, time.Millisecond)
	}
	<-time.After(time.Millisecond)
	assert.Equal(t, len(data.data), data.queue.Len())
	for i := 0; i < 20; i++ {
		cli.Set(fmt.Sprintf(KeyFmt, i), i, time.Millisecond)
	}
	<-time.After(time.Millisecond)
	cli.Set("a_15", 15, 0)
	assert.Equal(t, cli.Get("a_15"), 15)
}

func TestClient_Get_Extra(t *testing.T) {
	cli.Set("g0", 20, 0)
	data.freeze()
	assert.Equal(t, 20, cli.Get("g0"))
	cli.Del("g0")
}

func TestClient_Set_Extra(t *testing.T) {
	keyFmt := "s_%d"
	data.freeze()
	for i := 0; i < 20; i++ {
		cli.Set(fmt.Sprintf(keyFmt, i), 1, time.Millisecond)
	}
	data.toMove()
	//assert.Equal(t, 1, cli.Get("s_15"))
	cli.Set("s_19", 2, time.Millisecond)
	assert.Equal(t, 2, cli.Get("s_19"))

}

func TestClient_Del(t *testing.T) {
	key1 := "k_del1"
	key2 := "k_del2"
	key3 := "k_del3"
	cli.Set(key1, 1, 0)
	cli.Del(key1)
	assert.Equal(t, nil, cli.Get(key1))
	cli.Set(key1, 1, 0)
	cli.Set(key2, 2, 0)
	cli.Set(key3, 4, 0)
	data.freeze()
	cli.Set(key1, 3, 0)
	<-time.After(time.Millisecond)
	cli.Del(key1)
	assert.Equal(t, nil, cli.Get(key1))
	cli.Del(key2)
	assert.Equal(t, nil, cli.Get(key2))
	data.toMove()
	cli.Del(key1)
	assert.Equal(t, nil, cli.Get(key1))
	assert.Equal(t, 4, cli.Get(key3))
	cli.Del(key3)
	assert.Equal(t, nil, cli.Get(key3))
}

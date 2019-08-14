package model

import (
	"fmt"
	"net"
	"reflect"
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
)

func init() {
	Init(16)
}

func TestNewClient(t *testing.T) {
	type args struct {
		conn net.Conn
	}
	tests := []struct {
		name string
		args args
		want *Client
	}{
		{"new client", args{conn: nil}, &Client{Conn: nil, DataIdx: 0, Multi: &MultiInfo{false, false, nil, nil}}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewClient(tt.args.conn); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewClient() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestClient_Watch(t *testing.T) {
	const Num = 100
	cs := make([]*Client, Num)
	for i := range cs {
		cs[i] = NewClient(nil)
	}
	for i := 0; i < 100; i++ {
		cs[i].Watch(strconv.FormatInt(int64(i%5), 10))
		cs[i].Watch(strconv.FormatInt(int64(i%10), 10))
		if i%10 < 5 {
			cs[i].Multi.State = true
		}
	}
	assert.Equal(t, 10, len(data[0].watch))
	for i := 0; i < 5; i++ {
		k := strconv.FormatInt(int64(i), 10)
		assert.Equal(t, 20, data[0].watch[k].clients.Len())
		assert.Equal(t, k, data[0].watch[k].key)
	}
	for i := 5; i < 10; i++ {
		k := strconv.FormatInt(int64(i), 10)
		assert.Equal(t, 10, data[0].watch[k].clients.Len())
		assert.Equal(t, k, data[0].watch[k].key)
	}
	for i := 50; i < 100; i++ {
		cs[i].Unwatch()
	}
	cli := NewClient(nil)
	cli.Touch("10")
	cli.Touch("1")
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
	assert.Zero(t, len(data[0].watch))
}

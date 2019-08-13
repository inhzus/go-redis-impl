package model

import (
	"net"
	"reflect"
	"testing"
)

func TestNewClient(t *testing.T) {
	type args struct {
		conn net.Conn
	}
	tests := []struct {
		name string
		args args
		want *Client
	}{
		{"new client", args{conn: nil}, &Client{Conn: nil, DataIdx: 0, Multi: &MultiInfo{false, false, nil}}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewClient(tt.args.conn); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewClient() = %v, want %v", got, tt.want)
			}
		})
	}
}

package client

import (
	"fmt"
	"testing"

	"github.com/inhzus/go-redis-impl/internal/pkg/server"
	"github.com/stretchr/testify/assert"
)

var (
	c *Client
)

func init() {
	s := server.NewServer(&server.Option{})
	go func() {
		_ = s.Serve()
	}()
	c = NewClient(&Option{})
	_ = c.Connect()
}

func BenchmarkClient_Set(b *testing.B) {
	for i := 0; i < b.N; i++ {
		c.Set(fmt.Sprintf("A%d", i), i, 0)
	}
}

func BenchmarkClient_Pipeline(b *testing.B) {
	pipe := c.Pipeline()
	for i := b.N; i < b.N*2; i++ {
		pipe.Set(fmt.Sprintf("A%d", i), i, 0)
		if i%20 == 0 {
			_, err := pipe.Exec()
			assert.Nil(b, err)
		}
	}
}

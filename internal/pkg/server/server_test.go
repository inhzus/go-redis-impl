package server

import (
	"fmt"
	"strconv"
	"testing"
	"time"

	"github.com/inhzus/go-redis-impl/internal/pkg/client"
	"github.com/inhzus/go-redis-impl/internal/pkg/token"
	"github.com/stretchr/testify/assert"
)

var s *Server

func init() {
	s = NewServer(&Option{})
}

func TestServer_Serve(t *testing.T) {
	go func() { _ = s.Serve() }()
	<-time.After(time.Second / 10)
	cli := client.NewClient(&client.Option{})
	err := cli.Connect()
	if err != nil {
		t.Logf(err.Error())
		return
	}
	var res []*client.Response
	pipe := cli.Pipeline()
	for i := 0; i < 20; i++ {
		res = append(res, pipe.Set(fmt.Sprintf("a%d", i), i, 0))
	}
	_, err = pipe.Exec()
	if err != nil {
		t.Logf(err.Error())
		return
	}
	for _, r := range res {
		assert.Nil(t, r.Err, nil)
		assert.Equal(t, r.Data, token.ReplyOk)
	}
	res = nil
	for i := 0; i < 20; i++ {
		res = append(res, pipe.Get(fmt.Sprintf("a%d", i)))
	}
	_, err = pipe.Exec()
	if err == nil {
		for i := 0; i < 20; i++ {
			assert.Equal(t, []byte(strconv.FormatInt(int64(i), 10)), res[i].Data.Data)
		}
	} else {
		t.Logf(err.Error())
	}
	cli.Close()
	s.Close()
}

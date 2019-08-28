package client

import (
	"bytes"
	"fmt"
	"net"
	"time"

	"github.com/inhzus/go-redis-impl/internal/pkg/cds"
	"github.com/inhzus/go-redis-impl/internal/pkg/token"
)

// Response contains original tokens deserialized from binary raw
// and error generated during communicate with server.
type Response struct {
	Data *token.Token
	Err  error
}

// Task is the structure hold by the channel which connects front-end
// interactive functions and consumer.
type Task struct {
	Req []*token.Token
	Rsp chan *Response
}

// Option contains basic configurations of client.
type Option struct {
	Addr         string
	Database     int
	Proto        string
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
}

// Client is a go-redis-impl client and is safe for concurrent use
// by multiple goroutines.
type Client struct {
	option  *Option
	Conn    net.Conn
	queue   chan *Task
	stop    chan struct{}
	request func(*token.Token) *Response
}

// NewClient returns a pointer to Client entity with given option.
func NewClient(option *Option) *Client {
	if option.Proto == "" {
		option.Proto = "tcp"
	}
	if option.Proto == "unix" && option.Addr == "" {
		option.Addr = "/tmp/redis.sock"
	} else if option.Addr == "" {
		option.Addr = ":6389"
	}
	c := &Client{option: option}
	c.request = c.req
	return c
}

/*
consume is the consumer which consumes tasks sent by the server side.
1. Why consumer is necessary?
Go-redis-impl client is single-threaded, so consume with using golang channel
is the best implementation which I think that is suitable for the single-
threaded situation.
2. Efforts on `timeout`
net.Conn interface supplies the function `SetWriteDeadline` and
`SetReadDeadline`. `SetWriteDeadline` is alright for the case. However,
`SetReadDeadline` does not acts as what I expect. Anyway, if I use the
function to set a read deadline, the reading may interrupt when the data is
on the way from the server to the client, that is, next time when the client
send the new request to the server, it may recognize the last response as the
correspond response. Fatally the case will chain up all subsequent responses
parsed by the consumer. So instead I start a new goroutine to wait for and
parse the response when the client need to read from the connection. When it
is timeout, the consumer will stops waiting for the signal that response is
reached and continues to wait for another request sent to the consumer, but
the goroutine will still hang until the response is reached. In this way, I
manage to prevent from reading the last response.
 */
func (c *Client) consume(ts []*token.Token, rspCh chan *Response) {
	defer close(rspCh)
	buffer := &bytes.Buffer{}
	for _, t := range ts {
		d, err := t.Serialize()
		if err != nil {
			rspCh <- &Response{Err: err}
			return
		}
		buffer.Write(d)
	}
	data := buffer.Bytes()
	var err error
	if c.option.WriteTimeout > 0 {
		_ = c.Conn.SetWriteDeadline(time.Now().Add(c.option.WriteTimeout))
	}
	_, err = c.Conn.Write(data)
	if err != nil {
		rspCh <- &Response{Err: err}
		return
	}
	endCh := make(chan struct{})
	var responses []*Response
	go func() {
		rsp, _ := token.Deserialize(c.Conn)
		for _, t := range rsp {
			responses = append(responses, &Response{Data: t})
		}
		endCh <- struct{}{}
	}()
	if c.option.ReadTimeout > 0 {
		select {
		case <-endCh:
		case <-time.After(c.option.ReadTimeout):
			rspCh <- &Response{Err: fmt.Errorf("read data from connection timeout")}
			return
		}
	} else {
		<-endCh
	}
	for _, rsp := range responses {
		rspCh <- rsp
	}
	return
}

// Connect tries to dial server, starts a consumer goroutine,
// and selects the given database index directly.
func (c *Client) Connect() (err error) {
	c.Conn, err = net.DialTimeout(c.option.Proto, c.option.Addr, time.Second)
	if err != nil {
		return
	}
	c.queue = make(chan *Task)
	c.stop = make(chan struct{})

	go func() {
		for {
			select {
			case <-c.stop:
				c.stop <- struct{}{}
				return
			case t := <-c.queue:
				c.consume(t.Req, t.Rsp)
			}
		}
	}()
	if c.option.Database > 0 {
		c.Submit(token.NewArray(token.NewString(cds.Select), token.NewInteger(int64(c.option.Database))))
	}
	return
}

// Close interrupts the connection to server and stops the consumer.
func (c *Client) Close() {
	if c.Conn != nil {
		_ = c.Conn.Close()
		c.stop <- struct{}{}
		<-c.stop
	}
}

// Submit submits serialized tokens to consumer and returns a channel
// with response.
func (c *Client) Submit(t ...*token.Token) <-chan *Response {
	ch := make(chan *Response, 1)
	if c.Conn == nil {
		go func() { ch <- &Response{Err: fmt.Errorf("connection is nil")} }()
		return ch
	}
	c.queue <- &Task{Req: t, Rsp: ch}
	return ch
}

func (c *Client) req(t *token.Token) *Response {
	rsp := <-c.Submit(t)
	return rsp
}

// Pipeline returns a go-redis-impl pipeline. It takes usage of
// the connection and consumer of the original client.
func (c *Client) Pipeline() *Pipeline {
	p := &Pipeline{Client: *c}
	p.request = p.req
	p.Client.request = p.req
	return p
}

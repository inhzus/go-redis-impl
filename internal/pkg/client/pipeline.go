package client

import (
	"fmt"

	"github.com/golang/glog"
	"github.com/inhzus/go-redis-impl/internal/pkg/token"
)

// Pipeline shares the connection and consumer of the original
// client. Certainly it is safe for concurrent use of
// multiple goroutines.
type Pipeline struct {
	Client
	commands []*token.Token
}

func (p *Pipeline) req(t *token.Token) *Response {
	p.commands = append(p.commands, t)
	return &Response{Data: t}
}

// Exec executes all the commands stored. It ends with clearing
// the commands array.
func (p *Pipeline) Exec() ([]*token.Token, error) {
	if len(p.commands) == 0 {
		return nil, nil
	}
	defer func() { p.commands = nil }()
	if p.Client.Conn == nil {
		return nil, fmt.Errorf("connection is nil")
	}
	var responses []*Response
	for t := range p.Client.Submit(p.commands...) {
		if t.Err != nil {
			glog.Errorf("%s", t.Err.Error())
		}
		responses = append(responses, t)
	}
	if len(responses) == len(p.commands) {
		for i, item := range responses {
			p.commands[i].Label = item.Data.Label
			p.commands[i].Data = item.Data.Data
		}
		return p.commands, nil
	}
	return nil, responses[0].Err
}

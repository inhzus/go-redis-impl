package client

import (
	"fmt"

	"github.com/inhzus/go-redis-impl/internal/pkg/token"
)

type Pipeline struct {
	Client
	commands []*token.Token
}

func (p *Pipeline) req(t *token.Token) *Response {
	p.commands = append(p.commands, t)
	return &Response{Data: t}
}

func (p *Pipeline) Exec() ([]*token.Token, error) {
	if len(p.commands) == 0 {
		return nil, nil
	}
	if p.Client.conn == nil {
		defer p.Client.Close()
		if err := p.Client.Connect(); err != nil {
			return nil, fmt.Errorf("connection nil")
		}
	}
	var responses []*Response
	for t := range p.Client.Submit(p.commands...) {
		responses = append(responses, t)
	}
	if len(responses) == len(p.commands) {
		for i, item := range responses {
			p.commands[i].Label = item.Data.Label
			p.commands[i].Data = item.Data.Label
		}
		defer func() { p.commands = nil }()
		return p.commands, nil
	}
	return nil, responses[0].Err
}

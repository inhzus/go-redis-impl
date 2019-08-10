package client

import (
	"fmt"

	"github.com/inhzus/go-redis-impl/internal/pkg/label"
	"github.com/inhzus/go-redis-impl/internal/pkg/token"
)

type Pipeline struct {
	*Client
	commands []*token.Token
}

func (p *Pipeline) req(t *token.Token) *token.Response {
	p.commands = append(p.commands, t)
	return &token.Response{Data: t}
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
	rsp := p.Client.Submit(token.NewArray(p.commands...))
	if rsp.Err != nil {
		return nil, rsp.Err
	}
	if rsp.Data.Label == label.Error {
		return nil, fmt.Errorf(rsp.Data.Data.(string))
	}
	if rsp.Data.Label != label.Array {
		return nil, fmt.Errorf("pipeline response label not expected: %v", rsp.Data.Label)
	}
	data := rsp.Data.Data.([]*token.Token)
	if len(data) != len(p.commands) {
		return nil, fmt.Errorf("pipeline response sequences not equal to request")
	}
	for i, item := range data {
		p.commands[i].Label = item.Label
		p.commands[i].Data = item.Data
	}
	defer func() {
		p.commands = nil
	}()
	return p.commands, nil
}

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

func (p *Pipeline) req(t *token.Token) (*token.Token, error) {
	p.commands = append(p.commands, t)
	return t, nil
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
	ins, err := p.Client.Submit(token.NewArray(p.commands...))
	if err != nil {
		return nil, err
	}
	if ins.Label == label.Error {
		return nil, fmt.Errorf(ins.Data.(string))
	}
	if ins.Label != label.Array {
		return nil, fmt.Errorf("pipeline response label not expected: %v", ins.Label)
	}
	data := ins.Data.([]*token.Token)
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

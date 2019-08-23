package model

import (
	"github.com/inhzus/go-redis-impl/internal/pkg/task"
	"github.com/inhzus/go-redis-impl/internal/pkg/token"
)

type SetMsg struct {
	Idx int
	T   *token.Token
}

type ModTask struct {
	Cmd     string
	DataIdx int
	Rsp     chan error
}

type CmdTask struct {
	Cli *Client
	Req *token.Token
	Rsp chan *token.Token
}

func (m *SetMsg) Msg() task.Msg    { return m }
func (t *ModTask) Task() task.Task { return t }
func (t *CmdTask) Task() task.Task { return t }

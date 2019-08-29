package model

import (
	"github.com/inhzus/go-redis-impl/internal/pkg/task"
	"github.com/inhzus/go-redis-impl/internal/pkg/token"
)

// ModTask is the structure hold by the channel which sends task to control
// data persistence from server persistence goroutine to processor.
type ModTask struct {
	Cmd     string
	DataIdx int
	Rsp     chan error
}

// CmdTask is the structure hold by the channel which sends task to
// execute command from server client managers to processor.
type CmdTask struct {
	Cli *Client
	Req *token.Token
	Rsp chan *token.Token
}

// Task prevents type cast from interface to Task.
func (t *ModTask) Task() task.Task { return t }
func (t *CmdTask) Task() task.Task { return t }

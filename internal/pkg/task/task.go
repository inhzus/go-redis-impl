package task

import (
	"github.com/inhzus/go-redis-impl/internal/pkg/model"
	"github.com/inhzus/go-redis-impl/internal/pkg/token"
)

// Task is required to connect between handler & consumer
type Task interface {
	Task() Task
}

type ModTask struct {
	Cmd     string
	DataIdx int
	Rsp     chan error
}

type CmdTask struct {
	Cli *model.Client
	Req *token.Token
	Rsp chan *token.Token
}

func (t *ModTask) Task() Task { return t }
func (t *CmdTask) Task() Task { return t }

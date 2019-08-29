package proc

import (
	"github.com/inhzus/go-redis-impl/internal/pkg/task"
	"github.com/inhzus/go-redis-impl/internal/pkg/token"
)

// SetMsg is the structure hold by the channel which sends message from
// model clients to server persistence goroutine.
type SetMsg struct {
	Idx int
	T   *token.Token
}

// Msg prevents type cast from interface to Msg.
func (m *SetMsg) Msg() task.Msg { return m }

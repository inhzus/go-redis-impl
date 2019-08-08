package command

import (
	"time"

	"github.com/inhzus/go-redis-impl/internal/pkg/model"
	"github.com/inhzus/go-redis-impl/internal/pkg/token"
)

const (
	StringPing = "ping"
	StringPong = "pong"
	CmdGet     = "get"
	CmdSet     = "set"
)

func ping() *token.Token {
	return token.NewString(StringPong)
}

func processString(s string) *token.Token {
	switch s {
	case StringPing:
		return ping()
	}
	return token.ErrorDefault
}

func set(tokens []*token.Token) *token.Token {
	if len(tokens) < 2 {
		return token.ErrorDefault
	}
	key, value := tokens[0], tokens[1]
	if key.Label != token.LabelString {
		return token.ErrorDefault
	}
	if value.Label != token.LabelBulked {
		return token.ErrorDefault
	}
	model.Data.Set(key.Data.(string), value.Data, time.Now().Add(time.Minute))
	return token.ReplyOk
}

func get(tokens []*token.Token) *token.Token {
	if len(tokens) < 1 {
		return token.ErrorDefault
	}
	key := tokens[0]
	if key.Label != token.LabelString {
		return token.ErrorDefault
	}
	return &token.Token{Label: token.LabelBulked, Data: model.Data.Get(key.Data.(string))}
}

func processArray(tokens []*token.Token) *token.Token {
	if len(tokens) < 1 {
		return token.ErrorDefault
	}
	cmd := tokens[0]
	tokens = tokens[1:]
	if cmd.Label != token.LabelString {
		return token.ErrorDefault
	}
	switch cmd.Data.(string) {
	case CmdSet:
		return set(tokens)
	case CmdGet:
		return get(tokens)
	}
	return token.ErrorDefault
}

func Process(argument *token.Token) *token.Token {
	switch argument.Label {
	case token.LabelString:
		return processString(argument.Data.(string))
	case token.LabelArray:
		return processArray(argument.Data.([]*token.Token))
	}
	return token.ErrorDefault
}

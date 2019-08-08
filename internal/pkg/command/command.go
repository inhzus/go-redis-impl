package command

import (
	"github.com/inhzus/go-redis-impl/internal/pkg/token"
)

const (
	StringPing = "ping"
	StringPong = "pong"
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

func Process(argument *token.Token) *token.Token {
	switch argument.Label {
	case token.LabelString:
		return processString(argument.Data.(string))
	}
	return token.ErrorDefault
}

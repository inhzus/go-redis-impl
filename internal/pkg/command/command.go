package command

import (
	"time"

	"github.com/inhzus/go-redis-impl/internal/pkg/model"
	"github.com/inhzus/go-redis-impl/internal/pkg/token"
)

const (
	CmdGet     = "get"
	CmdSet     = "set"
	CmdPing    = "ping"
	StringPong = "pong"
)

func ping(_ []*token.Token) *token.Token {
	return token.NewString(StringPong)
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

func processCommand(cmd *token.Token, args []*token.Token) *token.Token {
	switch cmd.Data.(string) {
	case CmdGet:
		return get(args)
	case CmdSet:
		return set(args)
	case CmdPing:
		return ping(args)
	}
	return token.ErrorDefault
}

func Process(req *token.Token) *token.Token {
	data := req.Data.([]*token.Token)
	if len(data) == 0 {
		return token.ErrorDefault
	}
	cmd := data[0]
	if cmd.Label == token.LabelString {
		return processCommand(cmd, data[1:])
	}
	var rspData []*token.Token
	for _, item := range data {
		rspData = append(rspData, Process(item))
	}
	return token.NewArray(rspData...)
}

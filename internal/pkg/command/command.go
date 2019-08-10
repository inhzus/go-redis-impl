package command

import (
	"github.com/inhzus/go-redis-impl/internal/pkg/label"
	"github.com/inhzus/go-redis-impl/internal/pkg/model"
	"github.com/inhzus/go-redis-impl/internal/pkg/token"
)

const (
	CmdGet  = "get"
	CmdIncr = "incr"
	CmdSet  = "set"
	CmdPing = "ping"
)

const (
	strPong = "pong"
)

const (
	eStrMismatch = "type of %v is %v instead of %v"
	eStrArgMore  = "not enough arguments"
)

func ping(_ ...*token.Token) *token.Token {
	return token.NewString(strPong)
}

func set(tokens ...*token.Token) *token.Token {
	if len(tokens) < 2 {
		return token.NewError(eStrArgMore)
	}
	key, value := tokens[0], tokens[1]
	if et := checkKeyType(key); et != nil {
		return token.NewError(et.Error())
	}
	if et := checkType(value, "value", label.Bulked, label.Integer, label.String);
		et != nil {
		return token.NewError(et.Error())
	}
	model.Set(key.Data.(string), value.Data, nil)
	return token.ReplyOk
}

func get(tokens ...*token.Token) *token.Token {
	if len(tokens) < 1 {
		return token.NewError(eStrArgMore)
	}
	key := tokens[0]
	if et := checkKeyType(key); et != nil {
		return token.NewError(et.Error())
	}
	val := model.Get(key.Data.(string))
	if data, err := ItfToBulked(val); err != nil {
		return token.NewError(err.Error())
	} else {
		return token.NewBulked(data)
	}
}

func incr(tokens ...*token.Token) *token.Token {
	rsp := get(tokens...)
	if rsp.Label == label.Error {
		return rsp
	}
	num, err := ItfToInt(rsp.Data)
	if err != nil {
		return token.NewError(err.Error())
	}
	var val int64 = 1
	if num != nil {
		val = num.(int64) + 1
	}
	rsp = token.NewInteger(val)
	set(tokens[0], rsp)
	return rsp
}

func processCommand(cmd *token.Token, args []*token.Token) *token.Token {
	switch cmd.Data.(string) {
	case CmdGet:
		return get(args...)
	case CmdSet:
		return set(args...)
	case CmdPing:
		return ping(args...)
	case CmdIncr:
		return incr(args...)
	}
	return token.ErrorDefault
}

func Process(req *token.Token) *token.Token {
	if req == nil {
		return token.ErrorDefault
	}
	data := req.Data.([]*token.Token)
	if len(data) == 0 {
		return token.ErrorDefault
	}
	cmd := data[0]
	if cmd.Label == label.String {
		return processCommand(cmd, data[1:])
	}
	var rspData []*token.Token
	for _, item := range data {
		rspData = append(rspData, Process(item))
	}
	return token.NewArray(rspData...)
}

package command

import (
	"time"

	"github.com/inhzus/go-redis-impl/internal/pkg/label"
	"github.com/inhzus/go-redis-impl/internal/pkg/model"
	"github.com/inhzus/go-redis-impl/internal/pkg/token"
)

const (
	CmdDesc = "desc"
	CmdGet  = "get"
	CmdIncr = "incr"
	CmdSet  = "set"
	CmdPing = "ping"
)

const (
	strPong       = "pong"
	TimeoutSec    = "EX"
	TimeoutMilSec = "PX"
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
	if err := checkKeyType(key); err != nil {
		return token.NewError(err.Error())
	}
	if err := checkType(value, "value", label.Bulked, label.Integer, label.String);
		err != nil {
		return token.NewError(err.Error())
	}
	var timeout time.Duration
	for i := 2; i < len(tokens); i++ {
		if err := checkType(tokens[i], "argument", label.String); err != nil {
			return token.NewError(err.Error())
		}
		arg := tokens[i].Data.(string)
		switch arg {
		case TimeoutMilSec, TimeoutSec:
			i++
			if len(tokens) < i {
				return token.NewError("argument missing of %v", arg)
			}
			if err := checkType(tokens[i], "timeout", label.Integer); err != nil {
				return token.NewError(err.Error())
			}
			num := time.Duration(tokens[i].Data.(int64))
			switch arg {
			case TimeoutSec:
				timeout = num * time.Second
			case TimeoutMilSec:
				timeout = num * time.Millisecond
			}
		default:
			return token.NewError("argument not recognized")
		}
	}
	model.Set(key.Data.(string), value.Data, timeout)
	return token.ReplyOk
}

func get(tokens ...*token.Token) *token.Token {
	if len(tokens) < 1 {
		return token.NewError(eStrArgMore)
	}
	key := tokens[0]
	if err := checkKeyType(key); err != nil {
		return token.NewError(err.Error())
	}
	val := model.Get(key.Data.(string))
	if data, err := ItfToBulked(val); err != nil {
		return token.NewError(err.Error())
	} else {
		return token.NewBulked(data)
	}
}

func step(n int64, tokens []*token.Token) *token.Token {
	if len(tokens) < 1 {
		return token.NewError(eStrArgMore)
	}
	key := tokens[0]
	if err := checkKeyType(key); err != nil {
		return token.NewError(err.Error())
	}
	oldVal := model.Get(key.Data.(string))
	num, err := ItfToInt(oldVal)
	if err != nil {
		return token.NewError(err.Error())
	}
	if num != nil {
		n = num.(int64) + n
	}
	t := token.NewInteger(n)
	set(tokens[0], t)
	return t
}

func incr(tokens ...*token.Token) *token.Token {
	return step(1, tokens)
}

func desc(tokens ...*token.Token) *token.Token {
	return step(-1, tokens)
}

func processCommand(cmd *token.Token, args []*token.Token) *token.Token {
	switch cmd.Data.(string) {
	case CmdDesc:
		return desc(args...)
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

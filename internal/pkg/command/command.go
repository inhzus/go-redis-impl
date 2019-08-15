package command

import (
	"time"

	"github.com/inhzus/go-redis-impl/internal/pkg/label"
	"github.com/inhzus/go-redis-impl/internal/pkg/model"
	"github.com/inhzus/go-redis-impl/internal/pkg/token"
)

// command string
const (
	CmdDesc    = "desc"
	CmdDiscard = "discard"
	CmdExec    = "exec"
	CmdGet     = "get"
	CmdIncr    = "incr"
	CmdMulti   = "multi"
	CmdSet     = "set"
	CmdPing    = "ping"
	CmdUnwatch = "unwatch"
	CmdWatch   = "watch"
)

// argument string
const (
	TimeoutSec    = "EX"
	TimeoutMilSec = "PX"
)

const (
	strPong      = "pong"
	eStrMismatch = "type of %v is %v instead of %v"
	eStrArgMore  = "not enough arguments"
)

func ping(_ *model.Client, _ ...*token.Token) *token.Token {
	return token.NewString(strPong)
}

func set(cli *model.Client, tokens ...*token.Token) *token.Token {
	if len(tokens) < 2 {
		return token.NewError(eStrArgMore)
	}
	key, value := tokens[0], tokens[1]
	if err := checkKeyType(key); err != nil {
		return token.NewError(err.Error())
	}
	if err := checkType(value, "value", label.Bulked, label.Integer, label.String); err != nil {
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
			if len(tokens) < i+1 {
				return token.NewError("argument missing of %s", arg)
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
	cli.Set(key.Data.(string), value.Data, timeout)
	return token.ReplyOk
}

func get(cli *model.Client, tokens ...*token.Token) *token.Token {
	if len(tokens) < 1 {
		return token.NewError(eStrArgMore)
	}
	key := tokens[0]
	if err := checkKeyType(key); err != nil {
		return token.NewError(err.Error())
	}
	val := cli.Get(key.Data.(string))
	data, err := ItfToBulked(val)
	if err != nil {
		return token.NewError(err.Error())
	}
	return token.NewBulked(data)
}

func step(cli *model.Client, tokens []*token.Token, n int64) *token.Token {
	if len(tokens) < 1 {
		return token.NewError(eStrArgMore)
	}
	key := tokens[0]
	if err := checkKeyType(key); err != nil {
		return token.NewError(err.Error())
	}
	oldVal := cli.Get(key.Data.(string))
	num, err := ItfToInt(oldVal)
	if err != nil {
		return token.NewError(err.Error())
	}
	if num != nil {
		n = num.(int64) + n
	}
	t := token.NewInteger(n)
	set(cli, tokens[0], t)
	return t
}

func incr(cli *model.Client, tokens ...*token.Token) *token.Token {
	return step(cli, tokens, 1)
}

func desc(cli *model.Client, tokens ...*token.Token) *token.Token {
	return step(cli, tokens, -1)
}

func multi(cli *model.Client, _ ...*token.Token) *token.Token {
	if cli.Multi.State {
		return token.NewError("multi calls can not be nested")
	}
	cli.Multi.State = true
	cli.Multi.Dirty = false
	return token.ReplyOk
}

func exec(cli *model.Client, _ ...*token.Token) *token.Token {
	if !cli.Multi.State {
		return token.NewError("exec without multi")
	}
	var responses []*token.Token
	cli.Multi.State = false
	if !cli.Multi.Dirty {
		for _, t := range cli.Multi.Queue {
			rsp := Process(cli, t)
			responses = append(responses, rsp)
		}
	}
	cli.Multi.Queue = nil
	cli.Unwatch()
	return token.NewArray(responses...)
}

func discard(cli *model.Client, _ ...*token.Token) *token.Token {
	if !cli.Multi.State {
		return token.NewError("discard calls without multi")
	}
	cli.Multi.State = false
	cli.Multi.Queue = nil
	cli.Unwatch()
	return token.ReplyOk
}

func watch(cli *model.Client, tokens ...*token.Token) *token.Token {
	if len(tokens) < 1 {
		return token.NewError(eStrArgMore)
	}
	if cli.Multi.State {
		return token.NewError("watch inside multi is not allowed")
	}
	cli.Watch(tokens[0].Data.(string))
	return token.ReplyOk
}

func unwatch(cli *model.Client, tokens ...*token.Token) *token.Token {
	cli.Unwatch()
	return token.ReplyOk
}

// Process returns result of parsing request command and arguments
func Process(cli *model.Client, req *token.Token) *token.Token {
	data := req.Data.([]*token.Token)
	cmd, args := data[0], data[1:]
	if cli.Multi.State {
		switch cmd.Data.(string) {
		case CmdDiscard, CmdExec, CmdMulti, CmdWatch:
		default:
			cli.Multi.Queue = append(cli.Multi.Queue, req)
			return token.ReplyQueued
		}
	}
	switch cmd.Data.(string) {
	case CmdDesc:
		return desc(cli, args...)
	case CmdDiscard:
		return discard(cli, args...)
	case CmdExec:
		return exec(cli, args...)
	case CmdGet:
		return get(cli, args...)
	case CmdIncr:
		return incr(cli, args...)
	case CmdMulti:
		return multi(cli, args...)
	case CmdPing:
		return ping(cli, args...)
	case CmdSet:
		return set(cli, args...)
	case CmdWatch:
		return watch(cli, args...)
	case CmdUnwatch:
		return unwatch(cli, args...)
	}
	return token.NewError("unrecognized command")
}

package command

import (
	"net"
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

// Client entity: client information stored
type Client struct {
	Conn net.Conn
	// database index selected
	DataIdx int
	// transaction state, true if transaction started
	MultiState bool
	// transaction queue
	Queue []*token.Token
}

// NewClient returns a client selecting database 0, transaction state false
func NewClient(conn net.Conn) *Client {
	return &Client{Conn: conn, DataIdx: 0, MultiState: false}
}

func ping(_ *Client, _ ...*token.Token) *token.Token {
	return token.NewString(strPong)
}

func set(cli *Client, tokens ...*token.Token) *token.Token {
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
	model.Set(cli.DataIdx, key.Data.(string), value.Data, timeout)
	return token.ReplyOk
}

func get(cli *Client, tokens ...*token.Token) *token.Token {
	if len(tokens) < 1 {
		return token.NewError(eStrArgMore)
	}
	key := tokens[0]
	if err := checkKeyType(key); err != nil {
		return token.NewError(err.Error())
	}
	val := model.Get(cli.DataIdx, key.Data.(string))
	data, err := ItfToBulked(val)
	if err != nil {
		return token.NewError(err.Error())
	}
	return token.NewBulked(data)
}

func step(cli *Client, tokens []*token.Token, n int64) *token.Token {
	if len(tokens) < 1 {
		return token.NewError(eStrArgMore)
	}
	key := tokens[0]
	if err := checkKeyType(key); err != nil {
		return token.NewError(err.Error())
	}
	oldVal := model.Get(cli.DataIdx, key.Data.(string))
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

func incr(cli *Client, tokens ...*token.Token) *token.Token {
	return step(cli, tokens, 1)
}

func desc(cli *Client, tokens ...*token.Token) *token.Token {
	return step(cli, tokens, -1)
}

func multi(cli *Client, _ ...*token.Token) *token.Token {
	if cli.MultiState {
		return token.NewError("multi calls can not be nested")
	}
	cli.MultiState = true
	return token.ReplyOk
}

func exec(cli *Client, _ ...*token.Token) *token.Token {
	if !cli.MultiState {
		return token.NewError("exec without multi")
	}
	var responses []*token.Token
	for _, t := range cli.Queue {
		rsp := Process(cli, t)
		responses = append(responses, rsp)
	}
	return token.NewArray(responses...)
}

// Process returns result of parsing request command and arguments
func Process(cli *Client, req *token.Token) *token.Token {
	data := req.Data.([]*token.Token)
	cmd, args := data[0], data[1:]
	if cli.MultiState {
		switch cmd.Data.(string) {
		case CmdDiscard, CmdExec, CmdMulti, CmdWatch:
		default:
			cli.Queue = append(cli.Queue, req)
			return token.ReplyQueued
		}
	}
	switch cmd.Data.(string) {
	case CmdDesc:
		return desc(cli, args...)
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
	}
	return token.NewError("unrecognized command")
}

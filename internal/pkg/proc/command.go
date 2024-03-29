package proc

import (
	"fmt"
	"net"
	"time"

	"github.com/inhzus/go-redis-impl/internal/pkg/cds"
	"github.com/inhzus/go-redis-impl/internal/pkg/label"
	"github.com/inhzus/go-redis-impl/internal/pkg/model"
	"github.com/inhzus/go-redis-impl/internal/pkg/task"
	"github.com/inhzus/go-redis-impl/internal/pkg/token"
)

// proc string
const (
	ModFreeze = "freeze"
	ModMove   = "move"
)

const (
	strPong      = "pong"
	eStrMismatch = "type of %v is %v instead of %v"
	eStrArgMore  = "not enough arguments"
)

// Processor handles all the tasks sent from the connection handlers
// to the consumer.
type Processor struct {
	ctrlMap map[string]func(*model.Client, ...*token.Token) *token.Token
	data    []*model.DataStorage
	Msgs    struct {
		Set chan *SetMsg
	}
}

// NewProcessor returns a pointer to the processor which has initialized
// table-driven methods and data storage.
func NewProcessor(n int) *Processor {
	p := &Processor{}
	p.ctrlMap = map[string]func(*model.Client, ...*token.Token) *token.Token{
		cds.Desc:    p.desc,
		cds.Discard: p.discard,
		cds.Exec:    p.exec,
		cds.Get:     p.get,
		cds.Incr:    p.incr,
		cds.Multi:   p.multi,
		cds.Select:  p.sel,
		cds.Set:     p.set,
		cds.Ping:    p.ping,
		cds.Unwatch: p.unwatch,
		cds.Watch:   p.watch,
	}
	p.data = model.NewDataArray(n)
	p.Msgs.Set = make(chan *SetMsg)
	return p
}

// NewMockClient returns a new mock client without conn.
func (p *Processor) NewMockClient() *model.Client {
	return model.NewClient(nil, p.data[0])
}

// NewClient returns a new client collecting stat.
func (p *Processor) NewClient(conn net.Conn) *model.Client {
	return &model.Client{Conn: conn, Data: p.data[0], Multi: &model.MultiInfo{}, Stat: true}
}

// GenBin is a generator which yields every key-value pair of the original data.
func (p *Processor) GenBin(idx int, ch chan<- []byte) {
	defer close(ch)
	data := p.data[idx].GetOrigin()
	if len(data) == 0 {
		return
	}
	for k, v := range data {
		val, _ := ItfToBulked(v.Row)
		t := token.NewArray(token.NewString(cds.Set), token.NewString(k), token.NewBulked(val))
		if v.Expire > 0 {
			t.Data = append(t.Data.([]*token.Token), token.NewString(cds.ExpireAtNano), token.NewInteger(v.Expire))
		}
		d, _ := t.Serialize()
		ch <- d
	}
}

// execCmd returns result of parsing request command and arguments
func (p *Processor) execCmd(cli *model.Client, req *token.Token) (ret *token.Token) {
	if req == nil {
		return token.NewError("empty request")
	}
	data, ok := req.Data.([]*token.Token)
	if !ok || len(data) == 0 {
		return token.NewError("empty token")
	}
	cmd, args := data[0], data[1:]
	if err := checkType(cmd, "command", label.String); err != nil {
		return token.NewError(err.Error())
	}
	if cli.Multi.State {
		switch cmd.Data.(string) {
		case cds.Discard, cds.Exec, cds.Multi, cds.Watch:
		default:
			cli.Multi.Queue = append(cli.Multi.Queue, req)
			return token.ReplyQueued
		}
	}
	if proc, ok := p.ctrlMap[cmd.Data.(string)]; ok {
		ret = proc(cli, args...)
	} else {
		ret = token.NewError("unrecognized command")
	}
	if ret.Label == label.Error || !cli.Stat {
		return
	}
	switch cmd.Data.(string) {
	case cds.Incr, cds.Desc, cds.Set:
		ret.Flag |= token.FlagSet
	}
	return
}

func (p *Processor) execMod(cmd string, index int) error {
	switch cmd {
	case ModFreeze:
		return p.data[index].Freeze()
	case ModMove:
		return p.data[index].ToMove()
	}
	return fmt.Errorf("unrecognized command")
}

// Do handles all the tasks consumed by the server consumer.
func (p *Processor) Do(tsk task.Task) {
	switch t := tsk.(type) {
	case *model.CmdTask:
		rsp := p.execCmd(t.Cli, t.Req)
		t.Rsp <- rsp
		if rsp.Flag&token.FlagSet > 0 {
			p.Msgs.Set <- &SetMsg{t.Cli.Data.Idx(), t.Req}
		}
	case *model.ModTask:
		t.Rsp <- p.execMod(t.Cmd, t.DataIdx)
	}
}

func (p *Processor) ping(_ *model.Client, _ ...*token.Token) *token.Token {
	return token.NewString(strPong)
}

func (p *Processor) set(cli *model.Client, tokens ...*token.Token) *token.Token {
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
	var expire int64
	for i := 2; i < len(tokens); i++ {
		if tokens[i] == nil {
			return token.NewError("empty token")
		}
		if err := checkType(tokens[i], "argument", label.String); err != nil {
			return token.NewError(err.Error())
		}
		arg := tokens[i].Data.(string)
		switch arg {
		case cds.TimeoutMilSec, cds.TimeoutSec, cds.ExpireAtNano:
			i++
			if len(tokens) < i+1 {
				return token.NewError("argument missing of %s", arg)
			}
			if err := checkType(tokens[i], "timeout", label.Integer); err != nil {
				return token.NewError(err.Error())
			}
			num := tokens[i].Data.(int64)
			switch arg {
			case cds.TimeoutSec:
				expire = time.Now().Add(time.Duration(num) * time.Second).UnixNano()
			case cds.TimeoutMilSec:
				expire = time.Now().Add(time.Duration(num) * time.Millisecond).UnixNano()
			case cds.ExpireAtNano:
				expire = num
			}
		default:
			return token.NewError("argument not recognized")
		}
	}
	cli.Set(key.Data.(string), value.Data, expire)
	return token.ReplyOk
}

func (p *Processor) get(cli *model.Client, tokens ...*token.Token) *token.Token {
	if len(tokens) < 1 {
		return token.NewError(eStrArgMore)
	}
	key := tokens[0]
	if err := checkKeyType(key); err != nil {
		return token.NewError(err.Error())
	}
	val := cli.Get(key.Data.(string))
	data, _ := ItfToBulked(val)
	return token.NewBulked(data)
}

func (p *Processor) step(cli *model.Client, tokens []*token.Token, n int64) *token.Token {
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
	p.set(cli, tokens[0], t)
	return t
}

func (p *Processor) incr(cli *model.Client, tokens ...*token.Token) *token.Token {
	return p.step(cli, tokens, 1)
}

func (p *Processor) desc(cli *model.Client, tokens ...*token.Token) *token.Token {
	return p.step(cli, tokens, -1)
}

func (p *Processor) multi(cli *model.Client, _ ...*token.Token) *token.Token {
	if cli.Multi.State {
		return token.NewError("multi calls can not be nested")
	}
	cli.Multi.State = true
	cli.Multi.Dirty = false
	return token.ReplyOk
}

func (p *Processor) exec(cli *model.Client, _ ...*token.Token) *token.Token {
	if !cli.Multi.State {
		return token.NewError("exec without multi")
	}
	var responses []*token.Token
	cli.Multi.State = false
	if !cli.Multi.Dirty {
		for _, t := range cli.Multi.Queue {
			rsp := p.execCmd(cli, t)
			responses = append(responses, rsp)
		}
	}
	cli.Multi.Queue = nil
	cli.Unwatch()
	return token.NewArray(responses...)
}

func (p *Processor) discard(cli *model.Client, _ ...*token.Token) *token.Token {
	if !cli.Multi.State {
		return token.NewError("discard calls without multi")
	}
	cli.Multi.State = false
	cli.Multi.Queue = nil
	cli.Unwatch()
	return token.ReplyOk
}

func (p *Processor) watch(cli *model.Client, tokens ...*token.Token) *token.Token {
	if len(tokens) < 1 {
		return token.NewError(eStrArgMore)
	}
	if cli.Multi.State {
		return token.NewError("watch inside multi is not allowed")
	}
	if err := checkKeyType(tokens[0]); err != nil {
		return token.NewError(err.Error())
	}
	cli.Watch(tokens[0].Data.(string))
	return token.ReplyOk
}

func (p *Processor) unwatch(cli *model.Client, _ ...*token.Token) *token.Token {
	cli.Unwatch()
	return token.ReplyOk
}

// select
func (p *Processor) sel(cli *model.Client, tokens ...*token.Token) *token.Token {
	if len(tokens) < 1 {
		return token.NewError(eStrArgMore)
	}
	if err := checkType(tokens[0], "index", label.Integer); err != nil {
		return token.NewError(err.Error())
	}
	idx := int(tokens[0].Data.(int64))
	cli.Data = p.data[idx]
	return token.ReplyOk
}

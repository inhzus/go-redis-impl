package proc

import (
	"reflect"
	"testing"
	"time"

	"github.com/inhzus/go-redis-impl/internal/pkg/cds"
	"github.com/inhzus/go-redis-impl/internal/pkg/model"
	"github.com/inhzus/go-redis-impl/internal/pkg/token"
	"github.com/stretchr/testify/assert"
)

var cli *model.Client
var proc *Processor

func init() {
	proc = NewProcessor(16)
	cli = model.NewClient(nil, proc.data[0])
}

func TestProcessor_ping(t *testing.T) {
	type args struct {
		in0 *model.Client
		in1 []*token.Token
	}
	tests := []struct {
		name string
		args args
		want *token.Token
	}{
		{"ping", args{cli, nil}, token.NewString(strPong)},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := proc.ping(tt.args.in0, tt.args.in1...); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ping() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestProcessor_set(t *testing.T) {
	type args struct {
		cli    *model.Client
		tokens []*token.Token
	}
	tests := []struct {
		name string
		args args
		want *token.Token
	}{
		{"missing arguments",
			args{cli, []*token.Token{token.NewString("a")}},
			token.NewError(eStrArgMore)},
		{"key type error",
			args{cli, []*token.Token{token.NewInteger(1), token.NewInteger(2)}},
			token.NewError("type of key is integer instead of string")},
		{"value type error",
			args{cli, []*token.Token{token.NewString("a"), token.NewArray()}},
			token.NewError("type of value is array instead of bulked/integer/string")},
		{"argument missing",
			args{cli, []*token.Token{token.NewString("a"),
				token.NewString("b"), token.NewString("EX")}},
			token.NewError("argument missing of EX")},
		{"argument type error",
			args{cli, []*token.Token{token.NewString("a"), token.NewBulked([]byte("test")),
				token.NewString("EX"), token.NewString("PX")}},
			token.NewError("type of timeout is string instead of integer")},
		{"argument key error",
			args{cli, []*token.Token{token.NewString("a"), token.NewInteger(2),
				token.NewInteger(1), token.NewInteger(100)}},
			token.NewError("type of argument is integer instead of string")},
		{"argument key error",
			args{cli, []*token.Token{token.NewString("a"), token.NewInteger(2),
				token.NewString("test"), token.NewInteger(100)}},
			token.NewError("argument not recognized")},
		{"success",
			args{cli, []*token.Token{token.NewString("a"), token.NewInteger(3),
				token.NewString("PX"), token.NewInteger(100)}},
			token.ReplyOk},
		{"success",
			args{cli, []*token.Token{token.NewString("a"), token.NewInteger(3),
				token.NewString("EX"), token.NewInteger(1)}},
			token.ReplyOk},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := proc.set(tt.args.cli, tt.args.tokens...); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("set() = %v, want %v", got, tt.want)
			}
		})
	}
	proc.set(cli, token.NewString("a"), token.NewInteger(4),
		token.NewString("PT"), token.NewInteger(time.Now().Add(time.Millisecond).UnixNano()))
	assert.Equal(t, []byte("4"), proc.get(cli, token.NewString("a")).Data)
	<-time.After(time.Millisecond)
	assert.Equal(t, nil, proc.get(cli, token.NewString("a")).Data)
}

func TestProcessor_get(t *testing.T) {
	proc.set(cli, []*token.Token{token.NewString("a"), token.NewInteger(4)}...)
	type args struct {
		cli    *model.Client
		tokens []*token.Token
	}
	tests := []struct {
		name string
		args args
		want *token.Token
	}{
		{"missing argument",
			args{cli, []*token.Token{}},
			token.NewError("not enough arguments")},
		{"key type error",
			args{cli, []*token.Token{token.NewBulked([]byte("a"))}},
			token.NewError("type of key is bulked instead of string")},
		{"success",
			args{cli, []*token.Token{token.NewString("a")}},
			token.NewBulked([]byte("4"))},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := proc.get(tt.args.cli, tt.args.tokens...); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("get() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestProcessor_step(t *testing.T) {
	proc.set(cli, []*token.Token{token.NewString("b"), token.NewBulked([]byte("test"))}...)
	type args struct {
		cli    *model.Client
		tokens []*token.Token
		n      int64
	}
	tests := []struct {
		name string
		args args
		want *token.Token
	}{
		{"missing argument",
			args{cli, []*token.Token{}, 1},
			token.NewError("not enough arguments")},
		{"key type error",
			args{cli, []*token.Token{token.NewBulked([]byte("a"))}, 1},
			token.NewError("type of key is bulked instead of string")},
		{"value type error",
			args{cli, []*token.Token{token.NewString("b")}, 1},
			token.NewError("value (test) cannot cast to int")},
		{"success",
			args{cli, []*token.Token{token.NewString("c")}, 1},
			token.NewInteger(1)},
		{"success",
			args{cli, []*token.Token{token.NewString("c")}, -10},
			token.NewInteger(-9)},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := proc.step(tt.args.cli, tt.args.tokens, tt.args.n); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("step() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestProcessor_incr(t *testing.T) {
	type args struct {
		cli    *model.Client
		tokens []*token.Token
	}
	tests := []struct {
		name string
		args args
		want *token.Token
	}{
		{"success",
			args{cli, []*token.Token{token.NewString("t_incr")}},
			token.NewInteger(1)},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := proc.incr(tt.args.cli, tt.args.tokens...); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("incr() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestProcessor_desc(t *testing.T) {
	type args struct {
		cli    *model.Client
		tokens []*token.Token
	}
	tests := []struct {
		name string
		args args
		want *token.Token
	}{
		{"success",
			args{cli, []*token.Token{token.NewString("t_desc")}},
			token.NewInteger(-1)},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := proc.desc(tt.args.cli, tt.args.tokens...); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("desc() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestProcessor_multi(t *testing.T) {
	type args struct {
		cli *model.Client
		in1 []*token.Token
	}
	tests := []struct {
		name string
		args args
		want *token.Token
	}{
		{"enter multi",
			args{cli, []*token.Token{}},
			token.ReplyOk},
		{"enter multi",
			args{cli, []*token.Token{}},
			token.NewError("multi calls can not be nested")},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := proc.multi(tt.args.cli, tt.args.in1...); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("multi() = %v, want %v", got, tt.want)
			}
		})
	}
	assert.Equal(t, token.NewError("multi calls can not be nested"),
		proc.execCmd(cli, token.NewArray(token.NewString(cds.Multi))))
	proc.exec(cli)
}

func TestProcessor_exec(t *testing.T) {
	key := "t_exec"
	oldValue := "old value"
	newValue := "new value"
	c := model.NewClient(nil, proc.data[0])
	assert.Equal(t, token.NewError("exec without multi"), proc.exec(cli))
	assert.Equal(t, token.ReplyOk, proc.multi(cli))

	assert.Equal(t, token.ReplyQueued, proc.execCmd(cli,
		token.NewArray(token.NewString(cds.Get), token.NewString(key))))
	assert.Equal(t, token.ReplyQueued, proc.execCmd(cli,
		token.NewArray(token.NewString(cds.Set), token.NewString(key), token.NewString(newValue))))

	assert.Equal(t, token.ReplyOk, proc.execCmd(c,
		token.NewArray(token.NewString(cds.Set), token.NewString(key), token.NewString(oldValue))))
	assert.Equal(t, token.NewBulked([]byte(oldValue)), proc.execCmd(c,
		token.NewArray(token.NewString(cds.Get), token.NewString(key))))

	assert.Equal(t, token.ReplyQueued, proc.execCmd(cli,
		token.NewArray(token.NewString(cds.Get), token.NewString(key))))
	assert.Equal(t, token.NewArray(token.NewBulked([]byte(oldValue)), token.ReplyOk, token.NewBulked([]byte(newValue))),
		proc.execCmd(cli, token.NewArray(token.NewString(cds.Exec))))
	assert.Equal(t, token.ReplyOk, proc.execCmd(cli, token.NewArray(token.NewString(cds.Multi))))
	assert.Equal(t, token.NewArray(), proc.execCmd(cli, token.NewArray(token.NewString(cds.Exec))))
}

func TestProcessor_discard(t *testing.T) {
	assert.Equal(t, token.NewError("discard calls without multi"),
		proc.execCmd(cli, token.NewArray(token.NewString(cds.Discard))))
	assert.Equal(t, token.ReplyOk, proc.multi(cli))
	assert.Equal(t, token.ReplyQueued, proc.execCmd(cli, token.NewArray(token.NewString(cds.Get), token.NewString("a"))))
	assert.Equal(t, token.ReplyOk, proc.execCmd(cli, token.NewArray(token.NewString(cds.Discard))))
	assert.Equal(t, token.NewError("exec without multi"), proc.execCmd(cli, token.NewArray(token.NewString(cds.Exec))))
	assert.Equal(t, token.ReplyOk, proc.execCmd(cli, token.NewArray(token.NewString(cds.Multi))))
	assert.Equal(t, token.NewArray(), proc.execCmd(cli, token.NewArray(token.NewString(cds.Exec))))
}

func TestProcessor_watch(t *testing.T) {
	c := model.NewClient(nil, proc.data[0])
	key := "t_watch"
	assert.Equal(t, token.NewError(eStrArgMore),
		proc.execCmd(cli, token.NewArray(token.NewString(cds.Watch))))
	assert.Equal(t, token.NewError("type of key is bulked instead of string"),
		proc.execCmd(cli, token.NewArray(token.NewString(cds.Watch), token.NewBulked([]byte(key)))))
	assert.Equal(t, token.ReplyOk,
		proc.execCmd(cli, token.NewArray(token.NewString(cds.Watch), token.NewString(key))))
	assert.Equal(t, token.ReplyOk,
		proc.execCmd(c, token.NewArray(token.NewString(cds.Set), token.NewString(key), token.NewInteger(2))))
	assert.Equal(t, token.ReplyOk,
		proc.execCmd(cli, token.NewArray(token.NewString(cds.Multi))))
	assert.Equal(t, token.NewError("watch inside multi is not allowed"),
		proc.execCmd(cli, token.NewArray(token.NewString(cds.Watch), token.NewString(key))))
	assert.Equal(t, token.ReplyQueued,
		proc.execCmd(cli, token.NewArray(token.NewString(cds.Get), token.NewString(key))))
	assert.Equal(t, token.NewArray(token.NewBulked([]byte("2"))),
		proc.execCmd(cli, token.NewArray(token.NewString(cds.Exec))))
	assert.Zero(t, cli.Multi.Watched)
	assert.Equal(t, token.ReplyOk,
		proc.execCmd(cli, token.NewArray(token.NewString(cds.Multi))))
	assert.Equal(t, token.ReplyOk,
		proc.execCmd(c, token.NewArray(token.NewString(cds.Set), token.NewString(key), token.NewInteger(3))))
	assert.Equal(t, token.ReplyQueued,
		proc.execCmd(cli, token.NewArray(token.NewString(cds.Get), token.NewString(key))))
	assert.Equal(t, token.NewArray(token.NewBulked([]byte("3"))),
		proc.execCmd(cli, token.NewArray(token.NewString(cds.Exec))))
	assert.Equal(t, token.ReplyOk,
		proc.execCmd(cli, token.NewArray(token.NewString(cds.Watch), token.NewString(key))))
	assert.Equal(t, token.ReplyOk,
		proc.execCmd(cli, token.NewArray(token.NewString(cds.Multi))))
	assert.Equal(t, token.NewInteger(4),
		proc.execCmd(c, token.NewArray(token.NewString(cds.Incr), token.NewString(key))))
	assert.Equal(t, token.ReplyQueued,
		proc.execCmd(cli, token.NewArray(token.NewString(cds.Get), token.NewString(key))))
	assert.Equal(t, token.NewArray(),
		proc.execCmd(cli, token.NewArray(token.NewString(cds.Exec))))
}

func TestProcessor_unwatch(t *testing.T) {
	c := model.NewClient(nil, proc.data[0])
	key := "t_unwatch"
	assert.Equal(t, token.ReplyOk,
		proc.execCmd(cli, token.NewArray(token.NewString(cds.Watch), token.NewString(key))))
	assert.Equal(t, token.ReplyOk,
		proc.execCmd(cli, token.NewArray(token.NewString(cds.Unwatch))))
	assert.Equal(t, token.ReplyOk,
		proc.execCmd(cli, token.NewArray(token.NewString(cds.Multi))))
	assert.Equal(t, token.ReplyOk,
		proc.execCmd(c, token.NewArray(token.NewString(cds.Set), token.NewString(key), token.NewInteger(1))))
	assert.Equal(t, token.ReplyQueued,
		proc.execCmd(cli, token.NewArray(token.NewString(cds.Get), token.NewString(key))))
	assert.Equal(t, token.NewArray(token.NewBulked([]byte("1"))),
		proc.execCmd(cli, token.NewArray(token.NewString(cds.Exec))))

	assert.Equal(t, token.ReplyOk,
		proc.execCmd(cli, token.NewArray(token.NewString(cds.Watch), token.NewString(key))))
	assert.Equal(t, token.ReplyOk,
		proc.execCmd(cli, token.NewArray(token.NewString(cds.Multi))))
	assert.Equal(t, token.ReplyQueued,
		proc.execCmd(cli, token.NewArray(token.NewString(cds.Unwatch))))
	assert.Equal(t, token.ReplyQueued,
		proc.execCmd(cli, token.NewArray(token.NewString(cds.Get), token.NewString(key))))
	assert.Equal(t, token.ReplyOk,
		proc.execCmd(c, token.NewArray(token.NewString(cds.Set), token.NewString(key), token.NewInteger(2))))
	assert.Equal(t, token.NewArray(),
		proc.execCmd(cli, token.NewArray(token.NewString(cds.Exec))))
}

func TestProcessor_sel(t *testing.T) {
	key := "t_sel"
	c := model.NewClient(nil, proc.data[0])
	assert.Equal(t, token.NewError(eStrArgMore), proc.sel(cli))
	assert.Equal(t, token.NewError("type of index is string instead of integer"), proc.sel(cli, token.NewString("1")))
	assert.Equal(t, token.ReplyOk, proc.sel(cli, token.NewInteger(1)))
	assert.Equal(t, token.ReplyOk, proc.set(cli, token.NewString(key), token.NewString("value")))
	assert.Equal(t, token.NewBulked(nil), proc.get(c, token.NewString(key)))
	assert.Equal(t, token.ReplyOk, proc.sel(c, token.NewInteger(1)))
	assert.Equal(t, token.NewBulked([]byte("value")), proc.get(c, token.NewString(key)))
}

func TestProcessor_Exec(t *testing.T) {
	assert.Equal(t, token.NewError("empty request"),
		proc.execCmd(cli, nil))
	assert.Equal(t, token.NewError("empty token"),
		proc.execCmd(cli, token.NewArray()))
	assert.Equal(t, token.NewError("type of command is bulked instead of string"),
		proc.execCmd(cli, token.NewArray(token.NewBulked([]byte(cds.Get)))))
	assert.Equal(t, token.NewString(strPong),
		proc.execCmd(cli, token.NewArray(token.NewString(cds.Ping))))
	assert.Equal(t, token.NewInteger(-1),
		proc.execCmd(cli, token.NewArray(token.NewString(cds.Desc), token.NewString("t_process"))))
	assert.Equal(t, token.NewInteger(0),
		proc.execCmd(cli, token.NewArray(token.NewString(cds.Incr), token.NewString("t_process"))))
	assert.Equal(t, token.ReplyOk,
		proc.execCmd(cli, token.NewArray(token.NewString(cds.Select), token.NewInteger(2))))
	assert.Equal(t, proc.data[2], cli.Data)
	assert.Equal(t, token.NewError("unrecognized command"),
		proc.execCmd(cli, token.NewArray(token.NewString("unknown command"))))
}

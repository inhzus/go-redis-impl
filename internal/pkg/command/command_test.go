package command

import (
	"reflect"
	"testing"

	"github.com/inhzus/go-redis-impl/internal/pkg/model"
	"github.com/inhzus/go-redis-impl/internal/pkg/token"
	"github.com/stretchr/testify/assert"
)

var cli *model.Client

func init() {
	cli = model.NewClient(nil)
	model.Init(16)
}

func Test_ping(t *testing.T) {
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
			if got := ping(tt.args.in0, tt.args.in1...); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ping() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_set(t *testing.T) {
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
			if got := set(tt.args.cli, tt.args.tokens...); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("set() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_get(t *testing.T) {
	set(cli, []*token.Token{token.NewString("a"), token.NewInteger(4)}...)
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
			if got := get(tt.args.cli, tt.args.tokens...); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("get() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_step(t *testing.T) {

	set(cli, []*token.Token{token.NewString("b"), token.NewBulked([]byte("test"))}...)
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
			if got := step(tt.args.cli, tt.args.tokens, tt.args.n); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("step() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_incr(t *testing.T) {
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
			if got := incr(tt.args.cli, tt.args.tokens...); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("incr() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_desc(t *testing.T) {
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
			if got := desc(tt.args.cli, tt.args.tokens...); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("desc() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_multi(t *testing.T) {
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
			if got := multi(tt.args.cli, tt.args.in1...); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("multi() = %v, want %v", got, tt.want)
			}
		})
	}
	assert.Equal(t, token.NewError("multi calls can not be nested"),
		Process(cli, token.NewArray(token.NewString(CmdMulti))))
	exec(cli)
}

func Test_exec(t *testing.T) {
	key := "t_exec"
	oldValue := "old value"
	newValue := "new value"
	c := model.NewClient(nil)
	assert.Equal(t, token.NewError("exec without multi"), exec(cli))
	assert.Equal(t, token.ReplyOk, multi(cli))

	assert.Equal(t, token.ReplyQueued, Process(cli,
		token.NewArray(token.NewString(CmdGet), token.NewString(key))))
	assert.Equal(t, token.ReplyQueued, Process(cli,
		token.NewArray(token.NewString(CmdSet), token.NewString(key), token.NewString(newValue))))

	assert.Equal(t, token.ReplyOk, Process(c,
		token.NewArray(token.NewString(CmdSet), token.NewString(key), token.NewString(oldValue))))
	assert.Equal(t, token.NewBulked([]byte(oldValue)), Process(c,
		token.NewArray(token.NewString(CmdGet), token.NewString(key))))

	assert.Equal(t, token.ReplyQueued, Process(cli,
		token.NewArray(token.NewString(CmdGet), token.NewString(key))))
	assert.Equal(t, token.NewArray(token.NewBulked([]byte(oldValue)), token.ReplyOk, token.NewBulked([]byte(newValue))),
		Process(cli, token.NewArray(token.NewString(CmdExec))))
	assert.Equal(t, token.ReplyOk, Process(cli, token.NewArray(token.NewString(CmdMulti))))
	assert.Equal(t, token.NewArray(), Process(cli, token.NewArray(token.NewString(CmdExec))))
}

func Test_discard(t *testing.T) {
	assert.Equal(t, token.NewError("discard calls without multi"),
		Process(cli, token.NewArray(token.NewString(CmdDiscard))))
	assert.Equal(t, token.ReplyOk, multi(cli))
	assert.Equal(t, token.ReplyQueued, Process(cli, token.NewArray(token.NewString(CmdGet), token.NewString("a"))))
	assert.Equal(t, token.ReplyOk, Process(cli, token.NewArray(token.NewString(CmdDiscard))))
	assert.Equal(t, token.NewError("exec without multi"), Process(cli, token.NewArray(token.NewString(CmdExec))))
	assert.Equal(t, token.ReplyOk, Process(cli, token.NewArray(token.NewString(CmdMulti))))
	assert.Equal(t, token.NewArray(), Process(cli, token.NewArray(token.NewString(CmdExec))))
}

func TestProcess(t *testing.T) {
	assert.Equal(t, token.NewString(strPong),
		Process(cli, token.NewArray(token.NewString(CmdPing))))
	assert.Equal(t, token.NewInteger(-1),
		Process(cli, token.NewArray(token.NewString(CmdDesc), token.NewString("t_process"))))
	assert.Equal(t, token.NewInteger(0),
		Process(cli, token.NewArray(token.NewString(CmdIncr), token.NewString("t_process"))))
	assert.Equal(t, token.NewError("unrecognized command"),
		Process(cli, token.NewArray(token.NewString("unknown command"))))
}

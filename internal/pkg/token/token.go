package token

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/inhzus/go-redis-impl/internal/pkg/label"
)

var (
	ProtocolSeps       = []byte("\r\n")
	NilData            = []byte("-1\r\n")
	NilBulkedLen int64 = -1
	ReplyOk            = NewString("ok")
	ReplyQueued        = NewString("queued")
)

const (
	// value stored changed
	FlagSet = 1 << iota
)

type Token struct {
	Data  interface{}
	Flag  uint64
	Label byte
}

func (t *Token) Serialize() ([]byte, error) {
	const ErrorMsg = "cast t Data to %v error, Data: %v"
	if t == nil || t.Data == nil {
		data := []byte{label.Bulked}
		data = append(data, NilData...)
		return data, nil
	}
	data := []byte{t.Label}
	src := t.Data
	switch t.Label {
	case label.Array:
		array, ok := src.([]*Token)
		if !ok {
			return nil, fmt.Errorf(ErrorMsg, "array", src)
		}
		data = append(data, strconv.FormatInt(int64(len(array)), 10)...)
		data = append(data, ProtocolSeps...)
		for _, arg := range array {
			item, err := arg.Serialize()
			if err != nil {
				return nil, err
			}
			data = append(data, item...)
		}
	case label.Error, label.String:
		if s, ok := src.(string); ok {
			data = append(data, s...)
			data = append(data, ProtocolSeps...)
		} else {
			return nil, fmt.Errorf(ErrorMsg, "string", src)
		}
	case label.Integer:
		if num, ok := src.(int64); ok {
			data = append(data, strconv.FormatInt(num, 10)...)
			data = append(data, ProtocolSeps...)
		} else {
			return nil, fmt.Errorf(ErrorMsg, "integer", src)
		}
	case label.Bulked:
		val, ok := src.([]byte)
		if !ok {
			return nil, fmt.Errorf(ErrorMsg, "bulked string", src)
		}
		data = append(data, strconv.FormatInt(int64(len(val)), 10)...)
		data = append(data, ProtocolSeps...)
		data = append(data, val...)
		data = append(data, ProtocolSeps...)
	}
	return data, nil
}

func (t *Token) Format() (s string) {
	if row, err := t.Serialize(); err == nil {
		return strings.ReplaceAll(string(row), "\r\n", " ")
	}
	return
}

func (t *Token) Error() error {
	if t.Label == label.Error {
		return fmt.Errorf(t.Data.(string))
	}
	return nil
}

func (t *Token) Value() interface{} {
	return t.Data
}

func (t *Token) Equal(o *Token) bool {
	return t.Data == o.Data && t.Label == o.Label
}

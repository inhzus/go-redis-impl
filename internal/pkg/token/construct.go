package token

import (
	"fmt"

	"github.com/inhzus/go-redis-impl/internal/pkg/label"
)

func NewString(s string) *Token {
	return &Token{Label: label.String, Data: s}
}

func NewError(format string, a ...interface{}) *Token {
	if len(a) != 0 {
		format = fmt.Sprintf(format, a...)
	}
	return &Token{Label: label.Error, Data: format}
}

func NewInteger(num int64) *Token {
	return &Token{Label: label.Integer, Data: num}
}

func NewBulked(data interface{}) *Token {
	return &Token{Label: label.Bulked, Data: data}
}

func NewArray(tokens ...*Token) *Token {
	return &Token{Label: label.Array, Data: tokens}
}

func NewToken(v interface{}) *Token {
	switch r := v.(type) {
	case []byte:
		return NewBulked(r)
	case int64:
		return NewInteger(r)
	case string:
		return NewString(r)
	}
	return nil
}

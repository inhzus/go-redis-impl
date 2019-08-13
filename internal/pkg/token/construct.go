package token

import (
	"fmt"

	"github.com/inhzus/go-redis-impl/internal/pkg/label"
)

func NewString(s string) *Token {
	return &Token{label.String, s}
}

func NewError(format string, a ...interface{}) *Token {
	if len(a) != 0 {
		format = fmt.Sprintf(format, a...)
	}
	return &Token{label.Error, format}
}

func NewInteger(num int64) *Token {
	return &Token{label.Integer, num}
}

func NewBulked(data interface{}) *Token {
	return &Token{label.Bulked, data}
}

func NewArray(tokens ...*Token) *Token {
	return &Token{label.Array, tokens}
}

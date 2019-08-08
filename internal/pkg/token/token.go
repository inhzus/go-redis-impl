package token

import (
	"bufio"
	"bytes"
	"fmt"
	"net"
	"strconv"
	"strings"
)

const (
	LabelString  = '+'
	LabelError   = '-'
	LabelInteger = ':'
	LabelBulked  = '$'
	LabelArray   = '*'
)

var (
	ProtocolSeps       = []byte("\r\n")
	NilData            = []byte("-1\r\n")
	NilBulkedLen int64 = -1
	ErrorDefault       = NewError("error")
	ReplyOk            = NewString("ok")
)

type Task struct {
	Req *Token
	Rsp chan *Token
}

type Token struct {
	Label byte
	Data  interface{}
}

func NewString(s string) *Token {
	return &Token{LabelString, s}
}

func NewError(s string) *Token {
	return &Token{LabelError, s}
}

func NewInteger(num int64) *Token {
	return &Token{LabelInteger, num}
}

func NewBulked(data []byte) *Token {
	return &Token{LabelBulked, data}
}

func NewArray(tokens ...*Token) *Token {
	return &Token{LabelArray, tokens}
}

func readUntil(reader *bufio.Reader, seps []byte) (line []byte, err error) {
	for {
		var s []byte
		if s, err = reader.ReadBytes(seps[len(seps)-1]); err != nil {
			return
		}
		line = append(line, s...)
		if bytes.HasSuffix(line, seps) {
			return line[:len(line)-len(seps)], nil
		}
	}
}

func parseItem(reader *bufio.Reader) (*Token, error) {
	row, err := readUntil(reader, ProtocolSeps)
	if err != nil {
		return nil, err
	}
	if len(row) < 1 {
		return nil, fmt.Errorf("row empty")
	}
	label := row[0]
	row = row[1:]
	switch label {
	case LabelString, LabelError:
		return &Token{Label: label, Data: string(row)}, nil
	case LabelInteger:
		num, err := strconv.ParseInt(string(row), 10, 64)
		if err != nil {
			return nil, err
		}
		return &Token{Label: LabelInteger, Data: num}, nil
	case LabelBulked:
		n, err := strconv.ParseInt(string(row), 10, 64)
		if err != nil {
			return nil, err
		}
		if n == NilBulkedLen {
			return &Token{Label: LabelBulked, Data: nil}, nil
		}
		data := make([]byte, n)
		if _, err = reader.Read(data); err != nil {
			return nil, err
		}
		return &Token{Label: LabelBulked, Data: data}, nil
	case LabelArray:
		n, err := strconv.ParseInt(string(row), 10, 64)
		if err != nil {
			return nil, err
		}
		var i int64 = 0
		var tokens []*Token
		for ; i < n; i++ {
			token, err := parseItem(reader)
			if err != nil {
				return nil, err
			}
			tokens = append(tokens, token)
		}
		return &Token{Label: LabelArray, Data: tokens}, nil
	}
	return nil, nil
}

func Deserialize(conn net.Conn) (*Token, error) {
	reader := bufio.NewReader(conn)
	token, err := parseItem(reader)
	if err != nil {
		return nil, err
	}
	return token, nil
}

func (t *Token) Serialize() ([]byte, error) {
	const ErrorMsg = "cast t Data to %v error, Data: %v"
	if t == nil {
		data := []byte{LabelBulked}
		data = append(data, NilData...)
		return data, nil
	}
	data := []byte{t.Label}
	src := t.Data
	switch t.Label {
	case LabelArray:
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
	case LabelError, LabelString:
		if s, ok := src.(string); ok {
			data = append(data, s...)
			data = append(data, ProtocolSeps...)
		} else {
			return nil, fmt.Errorf(ErrorMsg, "string", src)
		}
	case LabelInteger:
		if num, ok := src.(int64); ok {
			data = append(data, strconv.FormatInt(num, 10)...)
			data = append(data, ProtocolSeps...)
		} else {
			return nil, fmt.Errorf(ErrorMsg, "integer", src)
		}
	case LabelBulked:
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

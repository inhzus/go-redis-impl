package token

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"net"
	"strconv"

	"github.com/golang/glog"
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
)

type Argument struct {
	label byte
	data  interface{}
}

func NewString(s string) *Argument {
	return &Argument{LabelString, s}
}

func NewError(s string) *Argument {
	return &Argument{LabelError, s}
}

func NewInteger(num int64) *Argument {
	return &Argument{LabelInteger, num}
}

func NewBulked(data []byte) *Argument {
	return &Argument{LabelBulked, data}
}

func NewArray(arguments []*Argument) *Argument {
	return &Argument{LabelArray, arguments}
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

func parseItem(reader *bufio.Reader) (*Argument, error) {
	row, err := readUntil(reader, ProtocolSeps)
	if err != nil {
		return nil, err
	}
	if len(row) < 1 {
		return nil, errors.New("row empty")
	}
	label := row[0]
	row = row[1:]
	switch label {
	case LabelString, LabelError:
		return &Argument{label: label, data: string(row),}, nil
	case LabelInteger:
		num, err := strconv.ParseInt(string(row), 10, 64)
		if err != nil {
			return nil, err
		}
		return &Argument{label: LabelInteger, data: num,}, nil
	case LabelBulked:
		n, err := strconv.ParseInt(string(row), 10, 64)
		if err != nil {
			return nil, err
		}
		if n == NilBulkedLen {
			return &Argument{label: LabelBulked, data: nil}, nil
		}
		data := make([]byte, n)
		if _, err = reader.Read(data); err != nil {
			return nil, err
		}
		return &Argument{label: LabelBulked, data: data,}, nil
	case LabelArray:
		n, err := strconv.ParseInt(string(row), 10, 64)
		if err != nil {
			return nil, err
		}
		var i int64 = 0
		var arguments []*Argument
		for ; i < n; i++ {
			argument, err := parseItem(reader)
			if err != nil {
				return nil, err
			}
			arguments = append(arguments, argument)
		}
		return &Argument{label: LabelArray, data: arguments,}, nil
	}
	return nil, nil
}

func Deserialize(conn net.Conn) (*Argument, error) {
	reader := bufio.NewReader(conn)
	argument, err := parseItem(reader)
	if err != nil {
		glog.Errorf("parse read: %v", err)
		return nil, err
	}
	glog.Infof("request token: %+v", argument)
	return argument, nil
}

func Serialize(argument *Argument) ([]byte, error) {
	const ErrorMsg = "cast token data to %v error, data: %v"
	if argument == nil {
		data := []byte{LabelBulked}
		data = append(data, NilData...)
		return data, nil
	}
	data := []byte{argument.label}
	src := argument.data
	switch argument.label {
	case LabelArray:
		array, ok := src.([]*Argument)
		if !ok {
			return nil, errors.New(fmt.Sprintf(ErrorMsg, "array", src))
		}
		data = append(data, strconv.FormatInt(int64(len(array)), 10)...)
		data = append(data, ProtocolSeps...)
		for _, arg := range array {
			item, err := Serialize(arg)
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
			return nil, errors.New(fmt.Sprintf(ErrorMsg, "string", src))
		}
	case LabelInteger:
		if num, ok := src.(int64); ok {
			data = append(data, strconv.FormatInt(num, 10)...)
			data = append(data, ProtocolSeps...)
		} else {
			return nil, errors.New(fmt.Sprintf(ErrorMsg, "integer", src))
		}
	case LabelBulked:
		val, ok := src.([]byte)
		if !ok {
			return nil, errors.New(fmt.Sprintf(ErrorMsg, "bulked string", src))
		}
		data = append(data, strconv.FormatInt(int64(len(val)), 10)...)
		data = append(data, ProtocolSeps...)
		data = append(data, val...)
		data = append(data, ProtocolSeps...)
	}
	return data, nil
}

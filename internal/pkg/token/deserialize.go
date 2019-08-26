package token

import (
	"bufio"
	"bytes"
	"fmt"
	"net"
	"strconv"

	"github.com/inhzus/go-redis-impl/internal/pkg/label"
	"github.com/inhzus/go-redis-impl/internal/pkg/task"
)

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
	sign := row[0]
	row = row[1:]
	switch sign {
	case label.String, label.Error:
		return &Token{Label: sign, Data: string(row)}, nil
	case label.Integer:
		num, err := strconv.ParseInt(string(row), 10, 64)
		if err != nil {
			return nil, err
		}
		return &Token{Label: sign, Data: num}, nil
	case label.Bulked:
		n, err := strconv.ParseInt(string(row), 10, 64)
		if err != nil {
			return nil, err
		}
		if n == NilBulkedLen {
			return &Token{Label: sign, Data: nil}, nil
		}
		data := make([]byte, n+2)
		if _, err = reader.Read(data); err != nil {
			return nil, err
		}
		return &Token{Label: sign, Data: data[:len(data)-2]}, nil
	case label.Array:
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
		return &Token{Label: label.Array, Data: tokens}, nil
	}
	return nil, fmt.Errorf("unrecognized label")
}

func Deserialize(conn net.Conn) ([]*Token, error) {
	reader := bufio.NewReader(conn)
	var ts []*Token
	for {
		t, err := parseItem(reader)
		if err != nil {
			return nil, err
		}
		ts = append(ts, t)
		if reader.Buffered() <= 0 {
			break
		}
	}
	return ts, nil
}

func DeserializeGenerator(reader *bufio.Reader) (ch chan *DeSerMsg) {
	ch = make(chan *DeSerMsg)
	go func() {
		defer close(ch)
		for {
			t, err := parseItem(reader)
			ch <- &DeSerMsg{T: t, Err: err}
			if err != nil {
				return
			}
			if reader.Buffered() <= 0 {
				return
			}
		}
	}()
	return
}

type DeSerMsg struct {
	T   *Token
	Err error
}

func (m *DeSerMsg) Msg() task.Msg { return m }

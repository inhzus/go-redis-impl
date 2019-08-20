package proc

import (
	"fmt"
	"strconv"

	"github.com/inhzus/go-redis-impl/internal/pkg/label"
	"github.com/inhzus/go-redis-impl/internal/pkg/token"
)

func checkType(t *token.Token, name string, types ...byte) error {
	matched := false
	for _, typ := range types {
		if t.Label == typ {
			matched = true
		}
	}
	if !matched {
		return fmt.Errorf(eStrMismatch, name, label.ToStr(t.Label),
			label.ToStr(types...))
	}
	return nil
}

func checkKeyType(t *token.Token) error {
	return checkType(t, "key", label.String)
}

// ItfToBulked converts interface bulked
func ItfToBulked(v interface{}) (interface{}, error) {
	if v == nil {
		return nil, nil
	}
	var data []byte
	switch v.(type) {
	case []byte:
		data = v.([]byte)
	case int:
		num := v.(int)
		data = []byte(strconv.FormatInt(int64(num), 10))
	case int64:
		num := v.(int64)
		data = []byte(strconv.FormatInt(num, 10))
	case string:
		data = []byte(v.(string))
	default:
		return nil, fmt.Errorf("value cannot cast to bulked")
	}
	return data, nil
}

// ItfToInt converts interface to int
func ItfToInt(v interface{}) (interface{}, error) {
	if v == nil {
		return nil, nil
	}
	errMsg := "value (%v) cannot cast to int"
	var data int64
	var err error
	switch v.(type) {
	case []byte:
		s := string(v.([]byte)[:])
		if data, err = strconv.ParseInt(s, 10, 64); err != nil {
			return nil, fmt.Errorf(errMsg, s)
		}
	case string:
		s := v.(string)
		if data, err = strconv.ParseInt(s, 10, 64); err != nil {
			return nil, fmt.Errorf(errMsg, s)
		}
	case int64:
		data = v.(int64)
	default:
		return nil, fmt.Errorf(errMsg, v)
	}
	return data, nil
}

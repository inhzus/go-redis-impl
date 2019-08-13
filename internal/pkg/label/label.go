package label

import "strings"

// protocol type label
const (
	String  Label = '+'
	Error   Label = '-'
	Integer Label = ':'
	Bulked  Label = '$'
	Array   Label = '*'
)

// Label alias byte
type Label = byte

// ToStr convert Label to string for error string usage
func ToStr(ls ...Label) string {
	length := len(ls)
	if length == 0 {
		return ""
	}
	var builder strings.Builder
	builder.Grow(length*6 - 1)
	isFirst := true
	for _, l := range ls {
		if isFirst {
			isFirst = false
		} else {
			builder.WriteByte('/')
		}
		switch l {
		case String:
			builder.WriteString("string")
		case Error:
			builder.WriteString("error")
		case Integer:
			builder.WriteString("integer")
		case Bulked:
			builder.WriteString("bulked")
		case Array:
			builder.WriteString("array")
		default:
			builder.WriteString("<unknown type>")
		}
	}
	return builder.String()
}

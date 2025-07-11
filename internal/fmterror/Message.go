package fmterror

import (
	"fmt"
	"strings"

	"go.llib.dev/testcase/pp"
)

type Message struct {
	Name    string
	Cause   string
	Message []any
	Values  []Value
}

type Value struct {
	Label string
	Value any
}

type Formatted string

func (m Message) String() string {
	var (
		format string
		args   []interface{}
	)
	if m.Name != "" {
		format += "[%s] "
		args = append(args, m.Name)
	}
	if m.Cause != "" {
		format += "%s"
		args = append(args, m.Cause)
	}
	if 0 < len(m.Message) {
		format += "\n%s"
		args = append(args, strings.TrimSpace(fmt.Sprintln(m.Message...)))
	}
	for _, v := range m.Values {
		var value string
		if raw, ok := v.Value.(Formatted); ok {
			value = string(raw)
		} else {
			value = pp.Format(v.Value)
		}
		format += "\n%s:"
		if 0 < strings.Count(value, "\n") {
			format += "\n\n%s\n"
		} else {
			format += "\t%s"
		}
		args = append(args, m.rightAlign(v.Label), value)
	}
	return fmt.Sprintf(format, args...)
}

func (m Message) rightAlign(str string) string {
	if strLen := len(str); strLen < m.labelLength() {
		str = strings.Repeat(" ", m.labelLength()-strLen) + str
	}
	return str
}

func (m Message) labelLength() int {
	var maxLength int
	for _, v := range m.Values {
		if length := len(v.Label); maxLength < length {
			maxLength = length
		}
	}
	return maxLength
}

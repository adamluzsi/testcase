package fmterror

import (
	"fmt"
	"strings"

	"github.com/adamluzsi/testcase/pp"
)

type Message struct {
	Method  string
	Cause   string
	Message []any
	Values  []Value
}

type Value struct {
	Label string
	Value interface{}
}

func (m Message) String() string {
	var (
		format string
		args   []interface{}
	)
	if m.Method != "" {
		format += "[%s] "
		args = append(args, m.Method)
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
		format += "\n%s:\t%s"
		args = append(args, m.rightAlign(v.Label), pp.Format(v.Value))
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

package assert

import (
	"fmt"
	"strings"
)

type message struct {
	Method      string
	Cause       string
	Left        *messageValue
	Right       *messageValue
	UserMessage []interface{}
}

type messageValue struct {
	Label string
	Value interface{}
}

func (m message) String() string {
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
	if 0 < len(m.UserMessage) {
		format += "\n%s"
		args = append(args, strings.TrimSpace(fmt.Sprintln(m.UserMessage...)))
	}
	if m.Left != nil {
		format += "\n%s:\t%#v"
		args = append(args, m.rightAlign(m.Left.Label), m.Left.Value)
	}
	if m.Right != nil {
		format += "\n%s:\t%#v"
		args = append(args, m.rightAlign(m.Right.Label), m.Right.Value)
	}
	return fmt.Sprintf(format, args...)
}

func (m message) rightAlign(str string) string {
	if strLen := len(str); strLen < m.labelLength() {
		str = strings.Repeat(" ", m.labelLength()-strLen) + str
	}
	return str
}

func (m message) labelLength() int {
	var maxLength int
	if m.Left != nil {
		maxLength = len(m.Left.Label)
	}
	if m.Right != nil && maxLength < len(m.Right.Label) {
		maxLength = len(m.Right.Label)
	}
	return maxLength
}

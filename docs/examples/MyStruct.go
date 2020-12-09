package examples

import (
	"fmt"
	"strings"
)

type MyStruct struct{}

func (ms MyStruct) Say() string {
	return `Hello, World!`
}

func (ms MyStruct) Foo() string {
	return `Foo`
}

func (ms MyStruct) Bar() string {
	return `Bar`
}

func (ms MyStruct) Baz() string {
	return `Baz`
}

func (ms MyStruct) Shrug(msg string) string {
	const shrugEmoji = `¯\_(ツ)_/¯`
	if !strings.HasSuffix(msg, shrugEmoji) {
		msg = fmt.Sprintf(`%s ¯\_(ツ)_/¯`, msg)
	}
	return msg
}

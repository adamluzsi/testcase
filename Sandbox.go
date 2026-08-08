package testcase

import (
	"go.llib.dev/testcase/sandbox"
)

func Sandbox(fn func()) sandbox.O {
	return sandbox.Run(fn)
}

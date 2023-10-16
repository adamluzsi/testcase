package testcase

import (
	"go.llib.dev/testcase/sandbox"
)

func Sandbox(fn func()) sandbox.RunOutcome {
	return sandbox.Run(fn)
}

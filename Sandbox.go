package testcase

import (
	"github.com/adamluzsi/testcase/sandbox"
)

func Sandbox(fn func()) sandbox.RunOutcome {
	return sandbox.Run(fn)
}

package internal

import (
	"testing"
)

func Verbose() bool {
	return verbose()
}

var verbose = testing.Verbose

func StubVerbose(tb testing.TB, fn func() bool) {
	tb.Cleanup(func() { verbose = testing.Verbose })
	verbose = fn
}

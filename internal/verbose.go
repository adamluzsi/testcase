package internal

import (
	"testing"
)

func Verbose() bool {
	return verbose()
}

var verbose = testing.Verbose

func StubVerbose[T bool | func() bool](tb testing.TB, v T) {
	tb.Cleanup(func() { verbose = testing.Verbose })
	switch v := any(v).(type) {
	case bool:
		verbose = func() bool { return v }
	case func() bool:
		verbose = v
	}
}

package assert

import (
	"sync"
	"testing"

	"github.com/adamluzsi/testcase/sandbox"

	"github.com/adamluzsi/testcase/internal/doubles"
	"github.com/adamluzsi/testcase/internal/fmterror"
)

// OneOf function checks a list of values and matches an expectation against each element of the list.
// If any of the elements pass the assertion, then the assertion helper function does not fail the test.
func OneOf[V any](tb testing.TB, vs []V, blk func(it It, got V), msg ...any) {
	tb.Helper()
	Must(tb).AnyOf(func(a *AnyOf) {
		a.name = "OneOf"
		a.cause = "None of the element matched the expectations"
		for _, v := range vs {
			a.Test(func(it It) { blk(it, v) })
			if a.OK() {
				break
			}
		}
	}, msg...)
}

// AnyOf is an assertion helper that allows you run AnyOf.Test assertion blocks, that can fail, as lone at least one of them succeeds.
// common usage use-cases:
//   - list of interface, where test order, or the underlying structure's implementation is irrelevant for the behavior.
//   - list of big structures, where not all field value relevant, only a subset, like a structure it wraps under a field.
//   - list of structures with fields that has dynamic state values, which is irrelevant for the given test.
//   - structure that can have various state scenario, and you want to check all of them, and you expect to find one match with the input.
//   - fan out scenario, where you need to check in parallel that at least one of the worker received the event.
type AnyOf struct {
	TB   testing.TB
	Fail func()

	mutex  sync.Mutex
	passed bool

	name  string
	cause string
}

// Test will test a block of assertion that must succeed in order to make AnyOf pass.
// You can have as much AnyOf.Test calls as you need, but if any of them pass with success, the rest will be skipped.
// Using Test is safe for concurrently.
func (ao *AnyOf) Test(blk func(t It)) {
	ao.TB.Helper()
	if ao.OK() {
		return
	}
	recorder := &doubles.RecorderTB{TB: ao.TB}
	defer recorder.CleanupNow()
	ro := sandbox.Run(func() {
		ao.TB.Helper()
		blk(MakeIt(recorder))
	})
	if !ro.Goexit && !ro.OK {
		ao.TB.Fatal("\n" + ro.Trace())
	}
	if recorder.IsFailed {
		return
	}
	ao.mutex.Lock()
	defer ao.mutex.Unlock()
	ao.passed = true
	return
}

// Finish will check if any of the assertion succeeded.
func (ao *AnyOf) Finish(msg ...interface{}) {
	ao.TB.Helper()
	if ao.OK() {
		return
	}
	ao.TB.Log(fmterror.Message{
		Method: func() string {
			if ao.name != "" {
				return ao.name
			}
			return "AnyOf"
		}(),
		Cause: func() string {
			if ao.cause != "" {
				return ao.cause
			}
			return "None of the .Test succeeded"
		}(),
		Message: msg,
		Values:  nil,
	})
	ao.Fail()
}

func (ao *AnyOf) OK() bool {
	ao.mutex.Lock()
	defer ao.mutex.Unlock()
	return ao.passed
}

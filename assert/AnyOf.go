package assert

import (
	"sync"
	"testing"

	"go.llib.dev/testcase/sandbox"

	"go.llib.dev/testcase/internal/doubles"
	"go.llib.dev/testcase/internal/fmterror"
)

// A stands for Any Of, an assertion helper that allows you run A.Case assertion blocks, that can fail, as lone at least one of them succeeds.
// common usage use-cases:
//   - list of interface, where test order, or the underlying structure's implementation is irrelevant for the behavior.
//   - list of big structures, where not all field value relevant, only a subset, like a structure it wraps under a field.
//   - list of structures with fields that has dynamic state values, which is irrelevant for the given test.
//   - structure that can have various state scenario, and you want to check all of them, and you expect to find one match with the input.
//   - fan out scenario, where you need to check in parallel that at least one of the worker received the event.
type A struct {
	TB   testing.TB
	Fail func()

	mutex  sync.Mutex
	passed bool

	name  string
	cause string
}

// Case will test a block of assertion that must succeed in order to make A pass.
// You can have as much A.Case calls as you need, but if any of them pass with success, the rest will be skipped.
// Using Case is safe for concurrently.
func (ao *A) Case(blk func(t It)) {
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
}

// Test is an alias for A.Case
func (ao *A) Test(blk func(t It)) {
	ao.TB.Helper()
	ao.Case(blk)
}

// Finish will check if any of the assertion succeeded.
func (ao *A) Finish(msg ...Message) {
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
		Message: toMsg(msg),
		Values:  nil,
	})
	ao.Fail()
}

func (ao *A) OK() bool {
	ao.mutex.Lock()
	defer ao.mutex.Unlock()
	return ao.passed
}

// OneOf function checks a list of values and matches an expectation against each element of the list.
// If any slice element meets the assertion, it is considered passed.
func OneOf[T any](tb testing.TB, vs []T, blk func(t It, got T), msg ...Message) {
	tb.Helper()
	Must(tb).AnyOf(func(a *A) {
		tb.Helper()
		a.name = "OneOf"
		a.cause = "None of the element matched the expectations"
		for _, v := range vs {
			a.Case(func(it It) {
				tb.Helper()

				blk(it, v)
			})
			if a.OK() {
				break
			}
		}
	}, msg...)
}

// NoneOf function checks a list of values and matches an expectation against each element of the list.
// If any slice element meets the assertion, it is considered failed.
func NoneOf[T any](tb testing.TB, vs []T, blk func(t It, got T), msg ...Message) {
	tb.Helper()

	var check = func(v T) bool {
		tb.Helper()
		dtb := &doubles.RecorderTB{TB: tb}
		sandbox.Run(func() {
			tb.Helper()
			blk(MakeIt(dtb), v)
		})

		assertFailed := dtb.IsFailed
		dtb.IsFailed = false // reset IsFailed for Cleanup

		sandbox.Run(func() {
			tb.Helper()
			dtb.CleanupNow()
		})
		if hasCleanupFailed := dtb.IsFailed; hasCleanupFailed {
			dtb.Forward()
		}

		return assertFailed
	}

	for i, v := range vs {
		if !check(v) {
			tb.Log(fmterror.Message{
				Method:  "NoneOf",
				Cause:   "One of the element matched the expectations",
				Message: toMsg(msg),
				Values: []fmterror.Value{
					{
						Label: "index",
						Value: i,
					},
					{
						Label: "value",
						Value: v,
					},
				},
			})
			tb.FailNow()
		}
	}
}

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
	TB testing.TB
	// FailWith [OPT] used to instruct assert.A on what to do when a failure occurs.
	// For example, if you want to skip a test is none of the assertion cases pass, then you can set testing.TB#SkipNow to FailWith
	//
	// Default: testing.TB#Fail
	FailWith func()
	// Name [OPT] will be used as the method/function name within the assertion failure message.
	Name string
	// Cause [OPT] will be used as a hint about what was the cause of the failure if the given `assert.A` fails`
	Cause string

	mutex      sync.Mutex
	passed     bool
	recordings []*doubles.RecorderTB
}

// Case will test a block of assertion that must succeed in order to make A pass.
// You can have as much A.Case calls as you need, but if any of them pass with success, the rest will be skipped.
// Using Case is safe for concurrently.
func (ao *A) Case(blk func(testing.TB)) {
	ao.TB.Helper()
	if ao.OK() {
		return
	}
	recorder := ao.newRecorder()
	defer recorder.CleanupNow()
	ro := sandbox.Run(func() {
		ao.TB.Helper()
		blk(recorder)
	})
	if !ro.Goexit && !ro.OK {
		ao.TB.Fatal("\n" + ro.Trace())
	}
	if recorder.IsFailed {
		return
	}
	ao.pass()
}

func (ao *A) newRecorder() *doubles.RecorderTB {
	ao.mutex.Lock()
	defer ao.mutex.Unlock()
	rtb := &doubles.RecorderTB{TB: ao.TB}
	ao.recordings = append(ao.recordings, rtb)
	return rtb
}

func (ao *A) pass() {
	ao.mutex.Lock()
	defer ao.mutex.Unlock()
	ao.passed = true
}

// Test is an alias for A.Case
func (ao *A) Test(blk func(t testing.TB)) {
	ao.TB.Helper()
	ao.Case(blk)
}

// Finish will check if any of the assertion succeeded.
func (ao *A) Finish(msg ...Message) {
	ao.TB.Helper()
	if ao.OK() {
		pass(ao.TB)
		return
	}

	ao.TB.Log(fmterror.Message{
		Name:    ao.Name,
		Cause:   ao.Cause,
		Message: toMsg(msg),
		Values:  nil,
	})

	if r, ok := ao.recordingOfTheMostLikelyCase(); ok {
		r.ForwardLogs()
	}

	ao.fail()
}

func (ao *A) fail() {
	if ao.FailWith != nil {
		ao.FailWith()
		return
	}

	ao.TB.Fail()
}

func (ao *A) recordingOfTheMostLikelyCase() (*doubles.RecorderTB, bool) {
	ao.mutex.Lock()
	defer ao.mutex.Unlock()

	if len(ao.recordings) == 0 {
		return nil, false
	}

	var (
		index = map[int][]int{}
		max   int
	)

	for i, r := range ao.recordings {
		var passes = r.Passes()
		if max < passes {
			max = passes
		}
		index[passes] = append(index[passes], i)
	}

	if is, ok := index[max]; ok && len(is) == 1 {
		return ao.recordings[is[0]], true
	}

	return nil, false
}

func (ao *A) OK() bool {
	ao.mutex.Lock()
	defer ao.mutex.Unlock()
	return ao.passed
}

// AnyOf is an assertion helper that deems the test successful
// if any of the declared assertion cases pass.
// This is commonly used when multiple valid formats are acceptable
// or when working with a list where any element meeting a certain criteria is considered sufficient.
func AnyOf(tb testing.TB, blk func(a *A), msg ...Message) {
	tb.Helper()
	Must(tb).AnyOf(blk, msg...)
}

// OneOf function checks a list of values and matches an expectation against each element of the list.
// If any slice element meets the assertion, it is considered passed.
func OneOf[T any](tb testing.TB, vs []T, blk func(t testing.TB, got T), msg ...Message) {
	tb.Helper()
	AnyOf(tb, func(a *A) {
		tb.Helper()
		a.Name = "OneOf"
		a.Cause = "None of the element matched the expectations"

		for _, v := range vs {
			a.Case(func(it testing.TB) {
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
func NoneOf[T any](tb testing.TB, vs []T, blk func(t testing.TB, got T), msg ...Message) {
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
				Name:    "NoneOf",
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
	pass(tb)
}

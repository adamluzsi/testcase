package assert

import (
	"testing"
	"time"

	"github.com/adamluzsi/testcase/sandbox"

	"github.com/adamluzsi/testcase/internal/doubles"
)

func EventuallyWithin[T time.Duration | int](durationOrCount T) Eventually {
	switch v := any(durationOrCount).(type) {
	case time.Duration:
		return Eventually{RetryStrategy: Waiter{Timeout: v}}
	case int:
		return Eventually{RetryStrategy: RetryCount(v)}
	default:
		panic("invalid usage")
	}
}

// Eventually Automatically retries operations whose failure is expected under certain defined conditions.
// This pattern enables fault-tolerance.
//
// A common scenario where using Eventually will benefit you is testing concurrent operations.
// Due to the nature of async operations, one might need to wait
// and observe the system with multiple tries before the outcome can be seen.
type Eventually struct{ RetryStrategy RetryStrategy }

type RetryStrategy interface {
	// While implements the retry strategy looping part.
	// Depending on the outcome of the condition,
	// the RetryStrategy can decide whether further iterations can be done or not
	While(condition func() bool)
}

type RetryStrategyFunc func(condition func() bool)

func (fn RetryStrategyFunc) While(condition func() bool) { fn(condition) }

// Assert will attempt to assert with the assertion function block multiple times until the expectations in the function body met.
// In case expectations are failed, it will retry the assertion block using the RetryStrategy.
// The last failed assertion results would be published to the received testing.TB.
// Calling multiple times the assertion function block content should be a safe and repeatable operation.
func (r Eventually) Assert(tb testing.TB, blk func(it It)) {
	tb.Helper()
	var lastRecorder *doubles.RecorderTB

	isFailed := tb.Failed()
	r.RetryStrategy.While(func() bool {
		tb.Helper()
		lastRecorder = &doubles.RecorderTB{TB: tb}
		ro := sandbox.Run(func() {
			tb.Helper()
			blk(MakeIt(lastRecorder))
		})
		if !ro.OK && !ro.Goexit { // when panic
			tb.Fatal("\n" + ro.Trace())
		}
		if lastRecorder.IsFailed {
			lastRecorder.CleanupNow()
		}
		if !isFailed && tb.Failed() {
			tb.Log("input testing.TB failed during Eventually.Assert, no more retry will be attempted")
			return false // if outer testing.TB failed during the assertion, no retry is expected
		}
		return lastRecorder.IsFailed
	})

	if lastRecorder != nil {
		lastRecorder.Forward()
	}
}

func RetryCount(times int) RetryStrategy {
	return RetryStrategyFunc(func(condition func() bool) {
		for i := 0; i < times+1; i++ {
			if ok := condition(); !ok {
				return
			}
		}
	})
}

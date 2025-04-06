package assert

import (
	"runtime"
	"testing"
	"time"

	"go.llib.dev/testcase/sandbox"

	"go.llib.dev/testcase/internal/doubles"
)

func MakeRetry[T time.Duration | int](durationOrCount T) Retry {
	switch v := any(durationOrCount).(type) {
	case time.Duration:
		return Retry{Strategy: Waiter{Timeout: v}}
	case int:
		return Retry{Strategy: RetryCount(v)}
	default:
		panic("impossible usage")
	}
}

// Retry Automatically retries operations whose failure is expected under certain defined conditions.
// This pattern enables fault-tolerance.
//
// A common scenario where using Retry will benefit you is testing concurrent operations.
// Due to the nature of async operations, one might need to wait
// and observe the system with multiple tries before the outcome can be seen.
type Retry struct{ Strategy RetryStrategy }

type RetryStrategy interface {
	// WaitWhile implements the retry strategy looping part.
	// Depending on the outcome of the condition,
	// the RetryStrategy can decide whether further iterations can be done or not
	WaitWhile(condition func() bool)
}

type RetryStrategyFunc func(condition func() bool)

func (fn RetryStrategyFunc) WaitWhile(condition func() bool) { fn(condition) }

// Assert will attempt to assert with the assertion function block multiple times until the expectations in the function body met.
// In case expectations are failed, it will retry the assertion block using the RetryStrategy.
// The last failed assertion results would be published to the received testing.TB.
// Calling multiple times the assertion function block content should be a safe and repeatable operation.
func (r Retry) Assert(tb testing.TB, blk func(t It)) {
	tb.Helper()
	var lastRecorder *doubles.RecorderTB

	isFailed := tb.Failed()
	r.Strategy.WaitWhile(func() bool {
		tb.Helper()
		runtime.Gosched()
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

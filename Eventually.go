package testcase

import (
	"testing"
	"time"

	"github.com/adamluzsi/testcase/assert"
	"github.com/adamluzsi/testcase/internal"
)

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
func (r Eventually) Assert(tb testing.TB, blk func(it assert.It)) {
	tb.Helper()
	var lastRecorder *internal.RecorderTB

	r.RetryStrategy.While(func() bool {
		tb.Helper()
		lastRecorder = &internal.RecorderTB{TB: tb}
		internal.RecoverExceptGoexit(func() {
			tb.Helper()
			blk(assert.MakeIt(lastRecorder))
		})
		if lastRecorder.IsFailed {
			lastRecorder.CleanupNow()
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

func makeEventually(i any) (Eventually, bool) {
	switch n := i.(type) {
	case time.Duration:
		return Eventually{RetryStrategy: Waiter{WaitTimeout: n}}, true
	case int:
		return Eventually{RetryStrategy: RetryCount(n)}, true
	case RetryStrategy:
		return Eventually{RetryStrategy: n}, true
	case Eventually:
		return n, true
	default:
		return Eventually{}, false
	}
}

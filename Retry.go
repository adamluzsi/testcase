package testcase

import (
	"github.com/adamluzsi/testcase/internal"
	"testing"
)

// Retry Automatically retries operations whose failure is expected under certain defined conditions.
// This pattern enables fault-tolerance.
//
// A common scenario where using Retry will benefit you is testing concurrent operations.
// Due to the nature of async operations, one might need to wait
// and observe the system with multiple tries before the outcome can be seen.
type Retry struct{ Strategy RetryStrategy }

type RetryStrategy interface {
	// While implements the retry strategy looping part.
	// Depending on the outcome of the condition,
	// the RetryStrategy can decide whether further iterations can be done or not
	While(condition func() bool)
}

// Assert will attempt to assert with the assertion function block multiple times until the expectations in the function body met.
// In case expectations are failed, it will retry the assertion block using the RetryStrategy.
// The last failed assertion results would be published to the received testing.TB.
// Calling multiple times the assertion function block content should be a safe and repeatable operation.
func (w Retry) Assert(tb testing.TB, blk func(testing.TB)) {
	var lastRecorder *internal.RecorderTB

	w.Strategy.While(func() bool {
		lastRecorder = &internal.RecorderTB{TB: tb}
		internal.InGoroutine(func() {
			blk(lastRecorder)
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

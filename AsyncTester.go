package testcase

import (
	"github.com/adamluzsi/testcase/internal"
	"testing"
)

type AsyncTester struct{ Waiter }

// Assert will attempt to assert with the assertion function block that expectations are met.
// In case expectations are failed, it will wait and attempt again to assert that the expectations are met.
// It behaves the same as WaitWhile, and if the wait timeout reached, the last failed assertion results would be published to the received testing.TB.
// Calling multiple times the assertion function block content should be a safe and repeatable operation.
func (w AsyncTester) Assert(tb testing.TB, blk func(testing.TB)) {
	var lastRecorder *internal.RecorderTB

	w.WaitWhile(func() bool {
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

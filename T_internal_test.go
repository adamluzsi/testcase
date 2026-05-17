package testcase

import (
	"sync/atomic"
	"testing"

	"go.llib.dev/testcase/internal/doubles"
)

// timerStub is a test double that implements both testing.TB and timerManager interfaces.
// It allows testing pauseTimer by tracking timer method calls.
type timerStub struct {
	testing.TB
	StopTimerCalled  int64
	StartTimerCalled int64
	ResetTimerCalled int64
}

func (m *timerStub) StopTimer() {
	atomic.AddInt64(&m.StopTimerCalled, 1)
}

func (m *timerStub) StartTimer() {
	atomic.AddInt64(&m.StartTimerCalled, 1)
}

func (m *timerStub) ResetTimer() {
	atomic.AddInt64(&m.ResetTimerCalled, 1)
}

func (m *timerStub) Helper() {}

func TestT_pauseTimer_withTimerManager(t *testing.T) {
	// Use a doubles.TB as the base for Helper() and other testing.TB methods
	baseTB := &doubles.TB{}
	stub := &timerStub{TB: baseTB}
	spec := NewSpec(stub)
	tcT := newT(stub, spec)

	// Call pauseTimer - should stop the timer
	resume := tcT.pauseTimer()

	// Verify StopTimer was called
	if atomic.LoadInt64(&stub.StopTimerCalled) != 1 {
		t.Errorf("expected StopTimer to be called once, got %d", stub.StopTimerCalled)
	}

	// Verify StartTimer has not been called yet
	if atomic.LoadInt64(&stub.StartTimerCalled) != 0 {
		t.Errorf("expected StartTimer to not be called yet, got %d", stub.StartTimerCalled)
	}

	// Call the resume function
	resume()

	// Verify StartTimer was called
	if atomic.LoadInt64(&stub.StartTimerCalled) != 1 {
		t.Errorf("expected StartTimer to be called once after resume, got %d", stub.StartTimerCalled)
	}
}

func TestT_pauseTimer_multipleNestedCalls(t *testing.T) {
	baseTB := &doubles.TB{}
	stub := &timerStub{TB: baseTB}
	spec := NewSpec(stub)
	tcT := newT(stub, spec)

	// First pause
	resume1 := tcT.pauseTimer()
	if atomic.LoadInt64(&stub.StopTimerCalled) != 1 {
		t.Errorf("expected StopTimer to be called once, got %d", stub.StopTimerCalled)
	}
	if atomic.LoadInt64(&stub.StartTimerCalled) != 0 {
		t.Errorf("expected StartTimer to not be called yet, got %d", stub.StartTimerCalled)
	}

	// Second nested pause - should not call StartTimer
	resume2 := tcT.pauseTimer()
	if atomic.LoadInt64(&stub.StopTimerCalled) != 2 {
		t.Errorf("expected StopTimer to be called twice, got %d", stub.StopTimerCalled)
	}
	if atomic.LoadInt64(&stub.StartTimerCalled) != 0 {
		t.Errorf("expected StartTimer to not be called yet after nested pause, got %d", stub.StartTimerCalled)
	}

	// First resume - should not call StartTimer because nested
	resume1()
	if atomic.LoadInt64(&stub.StartTimerCalled) != 0 {
		t.Errorf("expected StartTimer to not be called after first resume (nested), got %d", stub.StartTimerCalled)
	}

	// Second resume - should finally call StartTimer
	resume2()
	if atomic.LoadInt64(&stub.StartTimerCalled) != 1 {
		t.Errorf("expected StartTimer to be called once after all pauses resumed, got %d", stub.StartTimerCalled)
	}
}

func TestT_pauseTimer_withoutTimerManager(t *testing.T) {
	// Use a regular TB stub that doesn't implement timerManager
	regularStub := &doubles.TB{}
	spec := NewSpec(regularStub)
	tcT := newT(regularStub, spec)

	// Call pauseTimer - should return a no-op function
	resume := tcT.pauseTimer()

	// Calling resume should not panic or do anything
	resume()

	// The timerStopN should remain 0 since TB doesn't implement timerManager
	if tcT.timerStopN != 0 {
		t.Errorf("expected timerStopN to be 0, got %d", tcT.timerStopN)
	}
}

// TestT_pauseTimer_withNestedCalls verifies that nested pauseTimer calls
// correctly track the reference count, similar to what happens when var
// dependencies trigger nested timer pauses during initialization.
func TestT_pauseTimer_withNestedCalls(t *testing.T) {
	baseTB := &doubles.TB{}
	stub := &timerStub{TB: baseTB}
	spec := NewSpec(stub)
	tcT := newT(stub, spec)

	// Simulate what happens when a var's initialization triggers
	// another var's initialization, each calling pauseTimer:
	// 1. Outer operation starts (e.g., varB.Get) -> pauseTimer -> StopTimer #1
	resume1 := tcT.pauseTimer()

	// 2. Inner operation triggered (e.g., varA.Get from within varB) -> pauseTimer -> StopTimer #2
	resume2 := tcT.pauseTimer()

	// Both StopTimer calls should have been made
	if atomic.LoadInt64(&stub.StopTimerCalled) != 2 {
		t.Errorf("expected StopTimer to be called twice, got %d", stub.StopTimerCalled)
	}

	// timerStopN should reflect 2 active pauses
	if tcT.timerStopN != 2 {
		t.Errorf("expected timerStopN to be 2, got %d", tcT.timerStopN)
	}

	// When inner resumes first (defer order), timer should still be paused
	resume2()
	if tcT.timerStopN != 1 {
		t.Errorf("expected timerStopN to be 1 after first resume, got %d", tcT.timerStopN)
	}
	if atomic.LoadInt64(&stub.StartTimerCalled) != 0 {
		t.Errorf("expected StartTimer to not be called yet, got %d", stub.StartTimerCalled)
	}

	// When outer resumes, timer should finally start
	resume1()
	if tcT.timerStopN != 0 {
		t.Errorf("expected timerStopN to be 0 after all resumes, got %d", tcT.timerStopN)
	}
	if atomic.LoadInt64(&stub.StartTimerCalled) != 1 {
		t.Errorf("expected StartTimer to be called once, got %d", stub.StartTimerCalled)
	}
}

// TestT_pauseTimer_refCountDuringNestedExecution verifies that during
// nested pauseTimer calls, the timerStopN reference count is correctly maintained
// and StartTimer is only called when all nested pauses have been resumed.
func TestT_pauseTimer_refCountDuringNestedExecution(t *testing.T) {
	baseTB := &doubles.TB{}
	stub := &timerStub{TB: baseTB}
	spec := NewSpec(stub)
	tcT := newT(stub, spec)

	// Simulate nested pauseTimer calls like what happens during var dependency resolution
	// Outer "pause" (like varB init starting)
	outerResume := tcT.pauseTimer()

	// Inner "pause" (like varA init triggered by varB.Get)
	innerResume := tcT.pauseTimer()

	// At this point, StopTimer should be called twice
	if atomic.LoadInt64(&stub.StopTimerCalled) != 2 {
		t.Errorf("expected StopTimer to be called twice during nested pauses, got %d", stub.StopTimerCalled)
	}

	// timerStopN should be 2 while both are paused
	if tcT.timerStopN != 2 {
		t.Errorf("expected timerStopN to be 2 during nested pauses, got %d", tcT.timerStopN)
	}

	// Resume outer first - should NOT call StartTimer because inner is still paused
	outerResume()
	if atomic.LoadInt64(&stub.StartTimerCalled) != 0 {
		t.Errorf("expected StartTimer to not be called after outer resume (inner still paused), got %d", stub.StartTimerCalled)
	}

	// timerStopN should now be 1
	if tcT.timerStopN != 1 {
		t.Errorf("expected timerStopN to be 1 after outer resume, got %d", tcT.timerStopN)
	}

	// Resume inner - should finally call StartTimer
	innerResume()
	if atomic.LoadInt64(&stub.StartTimerCalled) != 1 {
		t.Errorf("expected StartTimer to be called after all resumes, got %d", stub.StartTimerCalled)
	}

	// timerStopN should now be 0
	if tcT.timerStopN != 0 {
		t.Errorf("expected timerStopN to be 0 after all resumes, got %d", tcT.timerStopN)
	}
}

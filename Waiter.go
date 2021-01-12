package testcase

import (
	"runtime"
	"time"
)

// Waiter is a component that waits for a time, event, or opportunity.
type Waiter struct {
	WaitDuration time.Duration
	WaitTimeout  time.Duration
}

// Wait will attempt to wait a bit and leave breathing space for other goroutines to steal processing time.
// It will also attempt to schedule other goroutines.
func (w Waiter) Wait() {
	finishTime := time.Now().Add(w.WaitDuration)
	for time.Now().Before(finishTime) {
		runtime.Gosched()
		time.Sleep(time.Nanosecond)
	}
}

// While will wait until a condition met, or until the wait timeout.
// By default, if the timeout is not defined, it just attempts to execute the condition once.
// Calling multiple times the condition function should be a safe operation.
func (w Waiter) While(condition func() bool) {
	finishTime := time.Now().Add(w.WaitTimeout)
	for condition() && time.Now().Before(finishTime) {
		w.Wait()
	}
}

package assert

import (
	"runtime"
	"time"
)

// Waiter is a component that waits for a time, event, or opportunity.
type Waiter struct {
	// WaitDuration is the time how lone Waiter.Wait should wait between attempting a new retry during Waiter.While.
	WaitDuration time.Duration
	// Timeout is used to calculate the deadline for the Waiter.While call.
	// If the retry takes longer than the Timeout, the retry will be cancelled.
	Timeout time.Duration
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
	finishTime := time.Now().Add(w.Timeout)
	for condition() && time.Now().Before(finishTime) {
		w.Wait()
	}
}

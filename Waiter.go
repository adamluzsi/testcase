package testcase

import (
	"runtime"
	"time"
)

// AsyncTester helps with asynchronous component testing.
// AsyncTester provides utility functionalities test scenarios where result is only expected to be eventually exist.
// A common testing scenario where using AsyncTester will benefit you is
// when the subject of test works with concurrency and returns earlier than when the result can be observed.
// Due to the nature of async operations, one might need to wait
// and observe the system with multiple tries before the outcome can be seen.
// By using AsyncTester for such testing use-cases,
// the testing should simplify by abstracting away the waiting and retrying related logic.
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

// WaitWhile will wait until a condition met, or until the wait timeout.
// By default, if the timeout is not defined, it just attempts to execute the condition once.
// Calling multiple times the condition function should be a safe operation.
func (w Waiter) WaitWhile(condition func() bool) {
	finishTime := time.Now().Add(w.WaitTimeout)
	for condition() && time.Now().Before(finishTime) {
		w.Wait()
	}
}

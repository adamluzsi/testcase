package testcase

import (
	"runtime"
	"sync"
	"testing"
	"time"
)

// Waiter can also mean someone/something who waits for a time, event, or opportunity.
// Waiter provides utility functionalities for waiting related test scenarios.
// The most common testing use-case to use Waiter when you need to test async operations related outcomes.
// Due to the nature of Async operations, one might need to wait and assert multiple times the outcome until the system processed a request.
// By using Waiter for such testing use-cases, the testing should simplify by abstracting away the waiting logic.
type Waiter struct {
	WaitDuration time.Duration
	WaitTimeout  time.Duration
}

// Wait will attempt to wait a bit and leave breathing space for other goroutines to steal processing time.
// It will also attempt to schedule other goroutines.
func (w Waiter) Wait() {
	times := runtime.NumGoroutine()
	sleepDuration := w.WaitDuration / time.Duration(times)
	for i := 0; i < times; i++ {
		runtime.Gosched()
		time.Sleep(sleepDuration)
	}
}

// WaitWhile will wait until a condition met, or until the wait timeout.
// By default, if the timeout is not defined, it just attempts to execute the condition once.
// Calling multiple times the condition function should be a safe operation.
func (w Waiter) WaitWhile(condition func() bool) {
	initialTime := time.Now()
	finishTime := initialTime.Add(w.WaitTimeout)
	for condition() && time.Now().Before(finishTime) {
		w.Wait()
	}
}

// Assert will attempt to assert with the assertion function block that expectations are met.
// In case expectations are failed, it will wait and attempt again to assert that the expectations are met.
// It behaves the same as WaitWhile, and if the wait timeout reached, the last failed assertion results would be published to the received testing.TB.
// Calling multiple times the assertion function block should be a safe operation.
func (w Waiter) Assert(tb testing.TB, assertionBlock func(testing.TB)) {
	var lastRecorder *recorderTB

	w.WaitWhile(func() bool {
		lastRecorder = &recorderTB{TB: tb}
		var wg sync.WaitGroup
		wg.Add(1)
		go func() {
			defer wg.Done()
			assertionBlock(lastRecorder)
		}()
		wg.Wait()
		return lastRecorder.isFailed
	})

	if lastRecorder != nil && lastRecorder.isFailed {
		lastRecorder.Replay(tb)
	}
}

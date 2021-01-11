package testcase_test

import (
	"math/rand"
	"testing"
	"time"

	"github.com/adamluzsi/testcase"
)

func ExampleAsyncTester_Assert() {
	waiter := testcase.Waiter{
		WaitDuration: time.Millisecond,
		WaitTimeout:  time.Second,
	}
	w := testcase.AsyncTester{Waiter: waiter}

	var t *testing.T
	// will attempt to wait until assertion block passes without a failing test result.
	// The maximum time it is willing to wait is equal to the wait timeout duration.
	// If the wait timeout reached, and there was no passing assertion run,
	// the last failed assertion history is replied to the received testing.TB
	//   In this case the failure would be replied to the *testing.T.
	w.Assert(t, func(tb testing.TB) {
		if rand.Intn(1) == 0 {
			tb.Fatal(`boom`)
		}
	})
}

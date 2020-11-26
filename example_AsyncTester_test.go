package testcase_test

import (
	"math/rand"
	"testing"
	"time"

	"github.com/adamluzsi/testcase"
)

func ExampleAsyncTester_Wait() {
	w := testcase.AsyncTester{
		WaitDuration: time.Millisecond,
	}

	w.Wait() // will wait 1 millisecond and attempt to schedule other go routines
}

func ExampleAsyncTester_WaitWhile() {
	w := testcase.AsyncTester{
		WaitDuration: time.Millisecond,
		WaitTimeout:  time.Second,
	}

	// will attempt to wait until condition returns false.
	// The maximum time it is willing to wait is equal to the wait timeout duration.
	w.WaitWhile(func() bool {
		return rand.Intn(1) == 0
	})
}

func ExampleAsyncTester_Assert() {
	w := testcase.AsyncTester{
		WaitDuration: time.Millisecond,
		WaitTimeout:  time.Second,
	}

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

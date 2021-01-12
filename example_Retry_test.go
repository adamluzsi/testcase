package testcase_test

import (
	"math/rand"
	"testing"
	"time"

	"github.com/adamluzsi/testcase"
)

func ExampleRetry_Assert() {
	waiter := testcase.Waiter{
		WaitDuration: time.Millisecond,
		WaitTimeout:  time.Second,
	}
	w := testcase.Retry{Strategy: waiter}

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

func ExampleRetry_asContextOption() {
	var tb testing.TB
	s := testcase.NewSpec(tb)

	s.Test(`flaky`, func(t *testcase.T) {

	}, testcase.Retry{Strategy: testcase.RetryCount(42)})
}

func ExampleRetryCount() {
	_ = testcase.Retry{Strategy: testcase.RetryCount(42)}
}

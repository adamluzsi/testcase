package testcase_test

import (
	"math/rand"
	"testing"
	"time"

	"github.com/adamluzsi/testcase"
	"github.com/adamluzsi/testcase/assert"
)

func ExampleEventually_Assert() {
	waiter := testcase.Waiter{
		WaitDuration: time.Millisecond,
		WaitTimeout:  time.Second,
	}
	w := testcase.Eventually{RetryStrategy: waiter}

	var t *testing.T
	// will attempt to wait until assertion block passes without a failing testCase result.
	// The maximum time it is willing to wait is equal to the wait timeout duration.
	// If the wait timeout reached, and there was no passing assertion run,
	// the last failed assertion history is replied to the received testing.TB
	//   In this case the failure would be replied to the *testing.T.
	w.Assert(t, func(it assert.It) {
		if rand.Intn(1) == 0 {
			it.Fatal(`boom`)
		}
	})
}

func ExampleEventually_asContextOption() {
	var tb testing.TB
	s := testcase.NewSpec(tb)

	s.Test(`flaky`, func(t *testcase.T) {
		// flaky test content here
	}, testcase.Flaky(testcase.RetryCount(42)))
}

func ExampleEventuallyCount() {
	_ = testcase.Eventually{RetryStrategy: testcase.RetryCount(42)}
}

func ExampleEventually_byTimeout() {
	r := testcase.Eventually{RetryStrategy: testcase.Waiter{
		WaitDuration: time.Millisecond,
		WaitTimeout:  time.Second,
	}}

	var t *testing.T
	r.Assert(t, func(it assert.It) {
		if rand.Intn(1) == 0 {
			it.Fatal(`boom`)
		}
	})
}

func ExampleEventually_byCount() {
	r := testcase.Eventually{RetryStrategy: testcase.RetryCount(42)}

	var t *testing.T
	r.Assert(t, func(it assert.It) {
		if rand.Intn(1) == 0 {
			it.Fatal(`boom`)
		}
	})
}

func ExampleEventually_byCustomRetryStrategy() {
	// this approach ideal if you need to deal with asynchronous systems
	// where you know that if a workflow process ended already,
	// there is no point in retrying anymore the assertion.

	while := func(isFailed func() bool) {
		for isFailed() {
			// just retry while assertion is failed
			// could be that assertion will be failed forever.
			// Make sure the assertion is not stuck in a infinite loop.
		}
	}

	r := testcase.Eventually{RetryStrategy: testcase.RetryStrategyFunc(while)}

	var t *testing.T
	r.Assert(t, func(it assert.It) {
		if rand.Intn(1) == 0 {
			it.Fatal(`boom`)
		}
	})
}

package testcase_test

import (
	"math/rand"
	"time"

	"github.com/adamluzsi/testcase"
)

func ExampleWaiter_Wait() {
	w := testcase.Waiter{WaitDuration: time.Millisecond}

	w.Wait() // will wait 1 millisecond and attempt to schedule other go routines
}

func ExampleWaiter_WaitWhile() {
	w := testcase.Waiter{
		WaitDuration: time.Millisecond,
		WaitTimeout:  time.Second,
	}

	// will attempt to wait until condition returns false.
	// The maximum time it is willing to wait is equal to the wait timeout duration.
	w.WaitWhile(func() bool {
		return rand.Intn(1) == 0
	})
}

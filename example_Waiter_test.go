package testcase_test

import (
	"math/rand"
	"time"

	"github.com/adamluzsi/testcase/assert"
)

func Example_assertWaiterWait() {
	w := assert.Waiter{WaitDuration: time.Millisecond}

	w.Wait() // will wait 1 millisecond and attempt to schedule other go routines
}

func Example_assertWaiterWhile() {
	w := assert.Waiter{
		WaitDuration: time.Millisecond,
		Timeout:      time.Second,
	}

	// will attempt to wait until condition returns false.
	// The maximum time it is willing to wait is equal to the wait timeout duration.
	w.While(func() bool {
		return rand.Intn(1) == 0
	})
}

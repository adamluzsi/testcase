//go:build !race
// +build !race

package testcase_test

// The build tag "race" is defined when building with the -race flag.

import (
	"fmt"
	"sync/atomic"
	"testing"
	"time"

	"go.llib.dev/testcase/sandbox"

	"go.llib.dev/testcase"
	"go.llib.dev/testcase/assert"
	"go.llib.dev/testcase/internal/doubles"
)

func TestRace(t *testing.T) {
	eventually := assert.Retry{Strategy: assert.Waiter{Timeout: time.Second}}

	t.Run(`functions run in race against each other`, func(t *testing.T) {
		eventually.Assert(t, func(it testing.TB) {
			var counter, total int32
			blk := func() {
				atomic.AddInt32(&total, 1)
				c := counter
				time.Sleep(time.Millisecond)
				counter = c + 1 // counter++ would not work
			}

			testcase.Race(blk, blk, blk, blk)
			assert.Equal(it, int32(4), total)
			it.Log(`counter:`, counter, `total:`, total)
			assert.True(it, counter < total,
				assert.Message(fmt.Sprintf(`counter was expected to be less that the total block run during race`)))
		})
	})

	t.Run(`each block runs once`, func(t *testing.T) {
		var sum int32
		testcase.Race(func() {
			atomic.AddInt32(&sum, 1)
		}, func() {
			atomic.AddInt32(&sum, 10)
		}, func() {
			atomic.AddInt32(&sum, 100)
		}, func() {
			atomic.AddInt32(&sum, 1000)
		})
		assert.Must(t).Equal(int32(1111), sum)
	})

	t.Run(`goexit propagated back from the lambdas after each lambda finished`, func(t *testing.T) {
		var fn1Finished, fn2Finished, afterRaceFinished bool
		sandbox.Run(func() {
			testcase.Race(func() {
				fn1Finished = true
			}, func() {
				fakeTB := &doubles.TB{}
				// this only meant to represent why goroutine exit needs to be propagated.
				fakeTB.FailNow()
				fn2Finished = true
			})
			afterRaceFinished = true
		})

		assert.Must(t).True(fn1Finished, `first race block was expected to finish regardless the second's FailNow call`)
		assert.Must(t).True(!fn2Finished, `second race block exited with FailNow, it shouldn't finished`)
		assert.Must(t).True(!afterRaceFinished, `after the second block exited, the exit should have propagated to the top one`)
	})
}

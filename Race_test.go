// +build !race

package testcase_test

// The build tag "race" is defined when building with the -race flag.

import (
	"fmt"
	"github.com/adamluzsi/testcase/internal"
	"sync/atomic"
	"testing"
	"time"

	"github.com/adamluzsi/testcase"
	"github.com/stretchr/testify/require"
)

func TestRace(t *testing.T) {
	retry := testcase.Retry{Strategy: testcase.Waiter{WaitTimeout: time.Second}}

	t.Run(`functions run in race against each other`, func(t *testing.T) {
		retry.Assert(t, func(tb testing.TB) {
			var counter, total int32
			blk := func() {
				atomic.AddInt32(&total, 1)
				c := counter
				time.Sleep(time.Millisecond)
				counter = c + 1 // counter++ would not work
			}

			testcase.Race(blk, blk, blk, blk)
			require.Equal(t, int32(4), total)
			tb.Log(`counter:`, counter, `total:`, total)
			require.True(t, counter < total,
				fmt.Sprintf(`counter was expected to be less that the total block run during race`))
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
		require.Equal(t, int32(1111), sum)
	})

	t.Run(`goexit propagated back from the lambdas after each lambda finished`, func(t *testing.T) {
		var fn1Finished, fn2Finished, afterRaceFinished bool
		internal.InGoroutine(func() {
			testcase.Race(func() {
				fn1Finished = true
			}, func() {
				fakeTB := &internal.StubTB{}
				// this only meant to represent why goroutine exit needs to be propagated.
				fakeTB.FailNow()
				fn2Finished = true
			})
			afterRaceFinished = true
		})

		require.True(t, fn1Finished, `first race block was expected to finish regardless the second's FailNow call`)
		require.False(t, fn2Finished, `second race block exited with FailNow, it shouldn't finished`)
		require.False(t, afterRaceFinished, `after the second block exited, the exit should have propagated to the top one`)
	})
}

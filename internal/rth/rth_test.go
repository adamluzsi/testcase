package rth_test

import (
	"context"
	"strconv"
	"testing"
	"time"

	"go.llib.dev/testcase/assert"
	"go.llib.dev/testcase/internal/rth"
	"go.llib.dev/testcase/random"
)

func TestSchedule(t *testing.T) {
	rnd := random.New(random.CryptoSeed{})
	done := make(chan struct{})
	defer close(done)

	for i := 0; i < 1000; i++ {
		// fire goroutines
		go func() {
			for {
				select {
				case <-done: // stop when the program is done
					return
				default: // do something that requires CPU time
					strconv.Itoa(rnd.Int())
				}
			}
		}()
	}

	var adjusted = func(d time.Duration, m float64) time.Duration {
		return time.Duration(float64(d) * m)
	}

	for _, dur := range []time.Duration{time.Second, time.Millisecond} {
		assert.Within(t, adjusted(dur, 1.2), func(ctx context.Context) {
			rth.Schedule(dur)
		})
	}
}

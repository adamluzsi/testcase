package wait_test

import (
	"context"
	"sync/atomic"
	"testing"
	"time"

	"go.llib.dev/testcase"
	"go.llib.dev/testcase/assert"
	"go.llib.dev/testcase/internal/wait"
	"go.llib.dev/testcase/pp"
)

func TestFor(t *testing.T) {
	s := testcase.NewSpec(t)

	s.Test("smoke", func(t *testcase.T) {
		done := make(chan struct{})
		defer close(done)

		var waitFor = t.Random.DurationBetween(time.Millisecond, time.Millisecond)
		var timeout = adjustDuration(waitFor, 1.3)
		t.OnFail(func() {
			t.Log("timeout", pp.Format(timeout))
			t.Log("waitFor", pp.Format(waitFor))
		})

		for i := 0; i < 1024; i++ {
			var pass int32
			assert.Within(t, timeout, func(ctx context.Context) {
				s := time.Now()
				wait.For(waitFor)
				d := time.Since(s)
				t.Cleanup(func() {
					if atomic.LoadInt32(&pass) != 1 {
						t.Log("time since", d.String())
					}
				})
				minDur := waitFor / 3
				gotDur := d
				assert.True(t, minDur <= gotDur,
					assert.MessageF("min(%s) <= got(%s)", minDur, gotDur))
			})
			atomic.SwapInt32(&pass, 1)
		}
	}, testcase.Flaky(3))
}

func TestOthers(t *testing.T) {
	done := make(chan struct{})
	defer close(done)

	var waitFor = time.Millisecond * 5
	var timeout = adjustDuration(waitFor, 1.3)
	t.Log("timeout", pp.Format(timeout))
	t.Log("waitFor", pp.Format(waitFor))

	for i := 0; i < 1024; i++ {
		s := time.Now()
		wait.Others(waitFor)
		d := time.Since(s)
		assert.True(t, d <= timeout)
	}
}

func adjustDuration(d time.Duration, m float64) time.Duration {
	return time.Duration(float64(d) * m)
}

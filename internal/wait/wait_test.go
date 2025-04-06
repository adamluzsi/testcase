package wait_test

import (
	"context"
	"sync/atomic"
	"testing"
	"time"

	"go.llib.dev/testcase/assert"
	"go.llib.dev/testcase/internal/wait"
	"go.llib.dev/testcase/pp"
)

func TestFor(t *testing.T) {
	done := make(chan struct{})
	defer close(done)

	var waitFor = time.Millisecond * 5
	var timeout = adjustDuration(waitFor, 1.3)
	t.Log("timeout", pp.Format(timeout))
	t.Log("waitFor", pp.Format(waitFor))

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
		})
		atomic.SwapInt32(&pass, 1)
	}
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

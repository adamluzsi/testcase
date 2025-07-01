package wait_test

import (
	"context"
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

		var waitFor = time.Millisecond
		var timeout = adjustDuration(waitFor, 1.4)

		for i := 0; i < 1024; i++ {
			assert.Within(t, timeout, func(ctx context.Context) {
				wait.For(waitFor)
			})
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

//go:build spike

package clock_test

import (
	"testing"
	"time"

	"go.llib.dev/testcase/clock"
	"go.llib.dev/testcase/clock/timecop"
)

func Test_spike_timeTicker(t *testing.T) {
	ticker := time.NewTicker(time.Second / 4)

	done := make(chan struct{})
	defer close(done)
	go func() {
		for {
			select {
			case <-ticker.C:
				t.Log("ticked")
			case <-done:
				return
			}
		}
	}()

	t.Log("4 expected on sleep")
	time.Sleep(time.Second + time.Microsecond)

	t.Log("now we reset and we expect 8 tick on sleep")
	ticker.Reset(time.Second / 8)
	time.Sleep(time.Second + time.Microsecond)

}

func Test_spike_clockTicker(t *testing.T) {
	ticker := clock.NewTicker(time.Second / 4)

	done := make(chan struct{})
	defer close(done)
	go func() {
		for {
			select {
			case <-ticker.C:
				t.Log("ticked")
			case <-done:
				return
			}
		}
	}()

	t.Log("4 expected on sleep")
	time.Sleep(time.Second + time.Microsecond)

	t.Log("now we reset and we expect 8 tick on sleep")
	ticker.Reset(time.Second / 8)
	time.Sleep(time.Second + time.Microsecond)

	t.Log("now time sped up, and where 4 would be expected on the following sleep, it should be 8")
	timecop.SetSpeed(t, 2)
	time.Sleep(time.Second + time.Microsecond)

}

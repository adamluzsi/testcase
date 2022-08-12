package clock

import (
	"github.com/adamluzsi/testcase/clock/internal"
	"time"
)

func TimeNow() time.Time {
	return internal.GetTime()
}

func Sleep(d time.Duration) {
	time.Sleep(internal.RemainingDuration(internal.GetTime(), d))
}

func After(d time.Duration) <-chan time.Time {
	startedAt := internal.GetTime()
	ch := make(chan time.Time)
	go func() {
	wait:
		for {
			duration := internal.RemainingDuration(startedAt, d)
			if duration <= 0 {
				break wait
			}
			select {
			case <-time.After(duration):
				break wait
			case <-internal.Listen(): // FIXME: flaky behaviour with time travelling
				continue wait
			}
		}
		ch <- TimeNow()
		close(ch)
	}()
	return ch
}

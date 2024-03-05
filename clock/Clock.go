package clock

import (
	"time"

	"go.llib.dev/testcase/clock/internal"
)

func TimeNow() time.Time {
	return internal.GetTime().Local()
}

func Sleep(d time.Duration) {
	<-After(d)
}

func After(d time.Duration) <-chan time.Time {
	ch := make(chan time.Time)
	if d == 0 {
		close(ch)
		return ch
	}
	startedAt := internal.GetTime()
	go func() {
	wait:
		for {
			select {
			case <-internal.Listen():
				continue wait
			case <-time.After(internal.RemainingDuration(startedAt, d)):
				break wait
			}
		}
		ch <- TimeNow()
		close(ch)
	}()
	return ch
}

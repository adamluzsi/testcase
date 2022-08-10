package clock

import (
	"github.com/adamluzsi/testcase/clock/internal"
	"time"
)

func TimeNow() time.Time {
	return time.Now().Add(internal.Chronos.Offset)
}

func Sleep(d time.Duration) {
	time.Sleep(internal.DurationFor(d))
}

func After(d time.Duration) <-chan time.Time {
	duration := internal.DurationFor(d)
	endTime := TimeNow().Add(duration)
	ch := make(chan time.Time)
	go func() {
		<-time.After(duration)
		ch <- endTime
	}()
	return ch
}

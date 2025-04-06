package wait

import (
	"runtime"
	"time"
)

func Others(timeout time.Duration) {
	const WaitUnit = time.Nanosecond
	var (
		goroutNum = runtime.NumGoroutine()
		startedAt = time.Now()
	)
	for i := 0; i < goroutNum; i++ { // since goroutines don't have guarantee when they will be scheduled
		runtime.Gosched() // we explicitly mark that we are okay with other goroutines to be scheduled
		elapsed := time.Since(startedAt)
		if timeout <= elapsed { // if max wait time is reached
			return
		}
		if elapsed < timeout { // if we withint the max wait time,
			time.Sleep(WaitUnit) // then we could just yield CPU too with sleep
		}
	}
}

func For(duration time.Duration) {
	if duration == 0 {
		runtime.Gosched()
		return
	}
	var (
		buffer  = duration / 8
		timeout = time.After(duration - buffer)
		grace   = time.After(duration / 2)
	)
waiting:
	for {
		select {
		case <-timeout:
			break waiting
		case <-grace:
			<-timeout
			break waiting
		default:
			runtime.Gosched()
		}
	}
	if buffer != 0 {
		time.Sleep(buffer)
	}
}

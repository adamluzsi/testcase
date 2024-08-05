package rth

import (
	"runtime"
	"time"
)

func Schedule(maxWait time.Duration) {
	const WaitUnit = time.Nanosecond
	var (
		goroutNum = runtime.NumGoroutine()
		startedAt = time.Now()
	)
	for i := 0; i < goroutNum; i++ { // since goroutines don't have guarantee when they will be scheduled
		runtime.Gosched() // we explicitly mark that we are okay with other goroutines to be scheduled
		elapsed := time.Since(startedAt)
		if maxWait <= elapsed { // if max wait time is reached
			return
		}
		if elapsed < maxWait { // if we withint the max wait time,
			time.Sleep(WaitUnit) // then we could just yield CPU too with sleep
		}
	}
}

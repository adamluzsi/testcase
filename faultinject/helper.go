package faultinject

import (
	"runtime"
	"time"
)

func wait() {
	for i, ngr := 0, runtime.NumGoroutine(); i < ngr*42; i++ {
		runtime.Gosched()
		time.Sleep(time.Nanosecond)
	}
}

package internal

import (
	"testing"
	"time"
)

var (
	NowFunc       func() time.Time
	SleepFunc     func(d time.Duration)
	AfterFunc     func(d time.Duration) <-chan time.Time
	NewTickerFunc func(d time.Duration) *TickerProxy
)

var _ = useTimeFunctions()

func init() {
	if testing.Testing() { // enable time travelling during testing
		useClockFunctions()
	}
}

func useTimeFunctions() struct{} {
	NowFunc = time.Now
	SleepFunc = time.Sleep
	AfterFunc = time.After
	NewTickerFunc = timeNewTicker
	return struct{}{}
}

func useClockFunctions() {
	NowFunc = func() time.Time {
		return TimeNow().Local()
	}
	SleepFunc = func(d time.Duration) {
		<-After(d)
	}
	AfterFunc = After
	NewTickerFunc = func(d time.Duration) *TickerProxy {
		ticker := NewTicker(d)
		return &TickerProxy{
			C:       ticker.C,
			onStop:  ticker.Stop,
			onReset: ticker.Reset,
		}
	}
}

package internal

import (
	"testing"
	"time"
)

var (
	NowFunc       func() time.Time
	SleepFunc     func(d time.Duration)
	AfterFunc     func(d time.Duration) <-chan time.Time
	SinceFunc     func(start time.Time) time.Duration
	NewTickerFunc func(d time.Duration) *Ticker
)

func init() {
	if testing.Testing() {
		useClockFunctions()
	} else {
		useTimeFunctions()
	}
}

func useTimeFunctions() struct{} {
	NowFunc = time.Now
	SleepFunc = time.Sleep
	AfterFunc = time.After
	NewTickerFunc = timeNewTicker
	SinceFunc = time.Since
	return struct{}{}
}

// useClockFunctions will enable the ability to time travel during testing.
func useClockFunctions() {
	NowFunc = Now
	SleepFunc = Sleep
	AfterFunc = After
	NewTickerFunc = NewTicker
	SinceFunc = Since
}

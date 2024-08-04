package clock

import (
	"time"

	"go.llib.dev/testcase/clock/internal"
)

// Now returns the current local time.
//
// During testing, Time returned by Now is affected by time travelling.
func Now() time.Time {
	return internal.NowFunc()
}

// Sleep pauses the current goroutine for at least the duration d.
// A negative or zero duration causes Sleep to return immediately.
//
// During testing, it will react to time travelling events
func Sleep(d time.Duration) {
	internal.SleepFunc(d)
}

// After waits for the duration to elapse and then sends the current time on the returned channel.
// The underlying Timer is not recovered by the garbage collector
//
// During testing, After will react to time travelling.
func After(d time.Duration) <-chan time.Time {
	return internal.After(d)
}

// NewTicker returns a new Ticker containing a channel that will send
// the current time on the channel after each tick. The period of the
// ticks is specified by the duration argument. The ticker will adjust
// the time interval or drop ticks to make up for slow receivers.
// The duration d must be greater than zero; if not, NewTicker will
// panic. Stop the ticker to release associated resources.
//
// During testing, Ticker will react to time travelling.
func NewTicker(d time.Duration) *Ticker {

	return internal.NewTickerFunc(d)
}

// Ticker acts as a proxy between the caller and the ticker implementation.
// During testing, it will be a clock-based ticker that can time travel,
// and outside of testing, it will use the time.Ticker.
type Ticker = internal.TickerProxy

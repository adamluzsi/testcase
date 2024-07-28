package clock

import (
	"sync"
	"time"

	"go.llib.dev/testcase/clock/internal"
)

// Now returns the current time.
// Time returned by Now is affected by time travelling.
func Now() time.Time {
	return internal.TimeNow().Local()
}

// TimeNow is an alias for clock.Now
func TimeNow() time.Time { return Now() }

func Sleep(d time.Duration) {
	<-After(d)
}

func After(d time.Duration) <-chan struct{} {
	startedAt := internal.TimeNow()
	ch := make(chan struct{})
	if d == 0 {
		close(ch)
		return ch
	}
	go func() {
		timeTravel := make(chan struct{})
		defer internal.Notify(timeTravel)()
		var onWait = func() (_restart bool) {
			c, td := after(internal.RemainingDuration(startedAt, d))
			defer td()
			select {
			case <-c:
				return false
			case <-timeTravel:
				return true
			}
		}
		for onWait() {
		}
		close(ch)
	}()
	return ch
}

func NewTicker(d time.Duration) *Ticker {
	ticker := &Ticker{d: d}
	ticker.init()
	return ticker
}

type Ticker struct {
	C chan time.Time

	d time.Duration

	onInit       sync.Once
	lock         sync.RWMutex
	done         chan struct{}
	pulse        chan struct{}
	ticker       *time.Ticker
	lastTickedAt time.Time
}

func (t *Ticker) init() {
	t.onInit.Do(func() {
		t.C = make(chan time.Time)
		t.done = make(chan struct{})
		t.pulse = make(chan struct{})
		t.ticker = time.NewTicker(t.getScaledDuration())
		t.updateLastTickedAt()
		go func() {
			timeTravel := make(chan struct{})
			defer internal.Notify(timeTravel)()
			for {
				if !t.ticking(timeTravel, t.ticker.C) {
					break
				}
			}
		}()
	})
}

func (t *Ticker) ticking(timeTravel <-chan struct{}, tick <-chan time.Time) bool {
	select {
	case <-t.done:
		return false

	case <-timeTravel: // on time travel, we reset the ticker according to the new time
		defer t.resetTicker()
		c, td := after(internal.RemainingDuration(t.getLastTickedAt(), t.getRealDuration()))
		defer td()
		return t.ticking(timeTravel, c) // wait the remaining time from the current tick

	case <-tick: // on timeout, we notify the listener
		now := t.updateLastTickedAt()
		t.C <- now
		return true
	}
}

// Stop turns off a ticker. After Stop, no more ticks will be sent.
// Stop does not close the channel, to prevent a concurrent goroutine
// reading from the channel from seeing an erroneous "tick".
func (t *Ticker) Stop() {
	t.lock.Lock()
	defer t.lock.Unlock()
	t.init()
	close(t.done)
	t.ticker.Stop()
	t.onInit = sync.Once{}
}

func (t *Ticker) Reset(d time.Duration) {
	t.init()
	t.setDuration(d)
	t.resetTicker()
}

func (t *Ticker) resetTicker() {
	d := t.getScaledDuration()
	if d == 0 { // zero is not an acceptable tick time
		d = time.Nanosecond
	}
	t.ticker.Reset(d)
}

// getScaledDuration returns the time duration that is altered by time
func (t *Ticker) getScaledDuration() time.Duration {
	return internal.ScaledDuration(t.getRealDuration())
}

func (t *Ticker) getRealDuration() time.Duration {
	t.lock.RLock()
	defer t.lock.RUnlock()
	return t.d
}

func (t *Ticker) setDuration(d time.Duration) {
	t.lock.Lock()
	defer t.lock.Unlock()
	t.d = d
}

func (t *Ticker) getLastTickedAt() time.Time {
	t.lock.RLock()
	defer t.lock.RUnlock()
	return t.lastTickedAt
}

func (t *Ticker) updateLastTickedAt() time.Time {
	t.lock.RLock()
	defer t.lock.RUnlock()
	t.lastTickedAt = Now()
	return t.lastTickedAt
}

func after(d time.Duration) (<-chan time.Time, func()) {
	if d == 0 {
		var ch = make(chan time.Time)
		close(ch)
		return ch, func() {}
	}
	timer := time.NewTimer(d)
	return timer.C, func() {
		if !timer.Stop() {
			select {
			case <-timer.C: // drain channel to unlock the resource
			default:
			}
		}
	}
}

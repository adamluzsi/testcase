package clock

import (
	"runtime"
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
		timeTravel := make(chan internal.TimeTravelEvent)
		defer internal.Notify(timeTravel)()
		defer close(ch)
		var handleTimeTravel func(tt internal.TimeTravelEvent) bool
		handleTimeTravel = func(tt internal.TimeTravelEvent) bool {
			deadline := startedAt.Add(d)
			if tt.When.After(deadline) || tt.When.Equal(deadline) {
				return true
			}
			if tt.Deep && tt.Freeze {
				// wait for next time travel, since during deep freeze, the flow of time is frozen
				return handleTimeTravel(<-timeTravel)
			}
			return false
		}
		if tt, ok := internal.Check(); ok && tt.Deep && tt.Freeze {
			if handleTimeTravel(tt) {
				return
			}
		}
		var onWait = func() (_restart bool) {
			c, td := after(internal.RemainingDuration(startedAt, d))
			defer td()
			select {
			case tt := <-timeTravel:
				return !handleTimeTravel(tt)
			case <-c:
				return false
			}
		}
		for onWait() {
		}
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
	lock         sync.RWMutex // is lock really needed if only the background goroutine reads the values from it?
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
			timeTravel := make(chan internal.TimeTravelEvent)
			defer internal.Notify(timeTravel)()

			if tt, ok := internal.Check(); ok { // trigger initial time travel awareness
				if !t.handleTimeTravel(timeTravel, tt) {
					return
				}
			}

			for {
				if !t.ticking(timeTravel, t.ticker.C, tickingOption{}) {
					break
				}
			}
		}()
	})
}

type tickingOption struct {
	// OnEvent will be executed when an event is received during waiting for ticking
	OnEvent func()
}

func (h tickingOption) onEvent() {
	if h.OnEvent == nil {
		return
	}
	h.OnEvent()
}

func (t *Ticker) ticking(timeTravel <-chan internal.TimeTravelEvent, tick <-chan time.Time, o tickingOption) bool {
	select {
	case <-t.done:
		o.onEvent()
		return false

	case tt := <-timeTravel: // on time travel, we reset the ticker according to the new time
		o.onEvent()
		return t.handleTimeTravel(timeTravel, tt)

	case <-tick: // on time.Ticker tick, we also tick
		o.onEvent()
		select {
		case tt := <-timeTravel:
			return t.handleTimeTravel(timeTravel, tt)
		case t.C <- t.updateLastTickedAt():
		}
		return true

	}
}

func (t *Ticker) handleTimeTravel(timeTravel <-chan internal.TimeTravelEvent, tt internal.TimeTravelEvent) bool {
	var (
		opt  = tickingOption{}
		prev = tt.Prev
		when = tt.When
	)
	if lastTickedAt := t.getLastTickedAt(); lastTickedAt.Before(prev) {
		prev = lastTickedAt
	}
	if fn := t.fastForwardTicksTo(prev, when); fn != nil {
		opt.OnEvent = fn
	}
	if tt.Deep && tt.Freeze {
		return t.ticking(timeTravel, nil, opt) // wait for unfreeze
	}
	defer t.resetTicker()
	c, td := after(internal.RemainingDuration(t.getLastTickedAt(), t.getRealDuration()))
	defer td()
	return t.ticking(timeTravel, c, opt) // wait the remaining time from the current tick
}

func (t *Ticker) fastForwardTicksTo(from, till time.Time) func() {
	var travelledDuration = till.Sub(from)

	if travelledDuration <= 0 {
		return nil
	}

	var (
		doneBeforeNextEvent = make(chan struct{})
		fastforwardWG       = &sync.WaitGroup{}
		timeBetweenTicks    = t.getRealDuration()
		missingTicks        = int(travelledDuration / timeBetweenTicks)
	)
	var OnBeforeEvent = func() {
		close(doneBeforeNextEvent)
		fastforwardWG.Wait()
	}

	// fast forward last ticked at position to the time after the ticks
	t.updateLastTickedAtTo(from.Add(timeBetweenTicks * time.Duration(missingTicks)))

	fastforwardWG.Add(1)
	go func(tickedAt time.Time) {
		defer fastforwardWG.Done()

	fastForward:
		for i := 0; i < missingTicks; i++ {
			tickedAt = tickedAt.Add(timeBetweenTicks) // move to the next tick time
			select {
			case <-doneBeforeNextEvent:
				break fastForward
			case t.C <- tickedAt: // tick!
				continue fastForward
			}
		}
	}(from)
	runtime.Gosched()

	return OnBeforeEvent
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
	return t.updateLastTickedAtTo(Now())
}

func (t *Ticker) updateLastTickedAtTo(at time.Time) time.Time {
	t.lock.RLock()
	defer t.lock.RUnlock()
	t.lastTickedAt = at
	return t.lastTickedAt
}

func after(d time.Duration) (<-chan time.Time, func()) {
	if d <= 0 {
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

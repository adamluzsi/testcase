package internal

import (
	"runtime"
	"sync"
	"time"
)

func timeNewTicker(d time.Duration) *Ticker {
	ticker := time.NewTicker(d)
	return &Ticker{
		C:       ticker.C,
		onStop:  ticker.Stop,
		onReset: ticker.Reset,
	}
}

func NewTicker(d time.Duration) *Ticker {
	ticker := NewTestTicker(d)
	return &Ticker{
		C:       ticker.C,
		onStop:  ticker.Stop,
		onReset: ticker.Reset,
	}
}

func Sleep(d time.Duration) {
	<-After(d)
}

func After(d time.Duration) <-chan time.Time {
	startedAt := Now()
	// The channel is buffered, so the send never blocks.
	// A receiver may stop listening, e.g. when another case of its select wins,
	// and an unbuffered send would then leak this goroutine with its time travel subscription.
	ch := make(chan time.Time, 1)
	if d == 0 {
		ch <- startedAt
		return ch
	}
	go func() {
		timeTravel := make(chan TimeTravelEvent)
		defer Notify(timeTravel)()
		defer close(ch)
		deadline := startedAt.Add(d)
		var handleTimeTravel = func(tt TimeTravelEvent) bool {
			for {
				if tt.When.After(deadline) || tt.When.Equal(deadline) {
					return true
				}
				if !(tt.Deep && tt.Freeze) {
					return false
				}
				// wait for next time travel, since during deep freeze, the flow of time is frozen
				tt = <-timeTravel
			}
		}
		if tt, ok := Check(); ok && tt.Deep && tt.Freeze {
			if handleTimeTravel(tt) {
				return
			}
		}
		var onWait = func() (_restart bool) {
			c, td := timeAfterWithCleanup(RemainingDuration(startedAt, d))
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
		ch <- Now()
	}()
	return ch
}

func timeAfterWithCleanup(d time.Duration) (<-chan time.Time, func()) {
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

// Ticker helps enable us to switch freely between time.Ticker and clock's Ticker implementation.
type Ticker struct {
	C <-chan time.Time

	onStop  func()
	onReset func(d time.Duration)
}

func (tp *Ticker) Stop() { tp.onStop() }

func (tp *Ticker) Reset(d time.Duration) { tp.onReset(d) }

func NewTestTicker(d time.Duration) *TestTicker {
	ticker := &TestTicker{duration: d}
	ticker.init()
	return ticker
}

type TestTicker struct {
	C chan time.Time

	duration     time.Duration
	onInit       sync.Once
	lock         sync.RWMutex // is lock really needed if only the background goroutine reads the values from it?
	done         chan struct{}
	pulse        chan struct{}
	ticker       *time.Ticker
	lastTickedAt time.Time
}

func (t *TestTicker) init() {
	t.onInit.Do(func() {
		t.C = make(chan time.Time)
		t.done = make(chan struct{})
		t.pulse = make(chan struct{})
		t.ticker = time.NewTicker(t.getScaledDuration())
		t.updateLastTickedAt()
		go func() {
			timeTravel := make(chan TimeTravelEvent)
			defer Notify(timeTravel)()

			w := t.regularTickWait()
			if tt, ok := Check(); ok { // trigger initial time travel awareness
				w = t.handleTimeTravel(tt)
			}

			// The ticker goroutine lives as long as the ticker,
			// so waits are chained in a loop rather than through recursion,
			// otherwise every time travel would grow its stack.
			for {
				var ok bool
				if w, ok = t.ticking(timeTravel, w); !ok {
					return
				}
			}
		}()
	})
}

// tickWait describes what the ticker goroutine waits for until its next event.
type tickWait struct {
	// tick signals the next tick.
	// A nil tick never fires, which is how a deep freeze halts the ticker.
	tick <-chan time.Time
	// OnEvent will be executed when an event is received during waiting for ticking
	OnEvent func()
	// OnEnd will be executed when the wait is over, after its event is handled.
	OnEnd func()
}

func (w tickWait) onEvent() {
	if w.OnEvent == nil {
		return
	}
	w.OnEvent()
}

func (w tickWait) onEnd() {
	if w.OnEnd == nil {
		return
	}
	w.OnEnd()
}

func (t *TestTicker) regularTickWait() tickWait {
	return tickWait{tick: t.ticker.C}
}

// ticking waits for the next event of w, and returns what to wait for next.
// It returns false when the ticker is stopped.
func (t *TestTicker) ticking(timeTravel <-chan TimeTravelEvent, w tickWait) (tickWait, bool) {
	defer w.onEnd()
	select {
	case <-t.done:
		w.onEvent()
		return tickWait{}, false

	case tt := <-timeTravel: // on time travel, we reset the ticker according to the new time
		w.onEvent()
		return t.handleTimeTravel(tt), true

	case <-w.tick: // on time.Ticker tick, we also tick
		w.onEvent()
		select {
		case tt := <-timeTravel:
			return t.handleTimeTravel(tt), true
		case t.C <- t.updateLastTickedAt():
		}
		return t.regularTickWait(), true

	}
}

// handleTimeTravel works out what the ticker should wait for after a time travel.
// It doesn't wait itself; ticking does.
func (t *TestTicker) handleTimeTravel(tt TimeTravelEvent) tickWait {
	var (
		w    tickWait
		prev = tt.Prev
		when = tt.When
	)
	if lastTickedAt := t.getLastTickedAt(); lastTickedAt.Before(prev) {
		prev = lastTickedAt
	}
	if fn := t.fastForwardTicksTo(prev, when); fn != nil {
		w.OnEvent = fn
	}
	if tt.Deep && tt.Freeze {
		return w // wait for unfreeze
	}
	c, td := timeAfterWithCleanup(RemainingDuration(t.getLastTickedAt(), t.getRealDuration()))
	w.tick = c // wait the remaining time from the current tick
	w.OnEnd = func() {
		td()
		t.resetTicker()
	}
	return w
}

func (t *TestTicker) fastForwardTicksTo(from, till time.Time) func() {
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
func (t *TestTicker) Stop() {
	t.lock.Lock()
	defer t.lock.Unlock()
	t.init()
	close(t.done)
	t.ticker.Stop()
	t.onInit = sync.Once{}
}

func (t *TestTicker) Reset(d time.Duration) {
	t.init()
	t.setDuration(d)
	t.resetTicker()
}

func (t *TestTicker) resetTicker() {
	d := t.getScaledDuration()
	if d == 0 { // zero is not an acceptable tick time
		d = time.Nanosecond
	}
	t.ticker.Reset(d)
}

// getScaledDuration returns the time duration that is altered by time
func (t *TestTicker) getScaledDuration() time.Duration {
	return ScaledDuration(t.getRealDuration())
}

func (t *TestTicker) getRealDuration() time.Duration {
	t.lock.RLock()
	defer t.lock.RUnlock()
	return t.duration
}

func (t *TestTicker) setDuration(d time.Duration) {
	t.lock.Lock()
	defer t.lock.Unlock()
	t.duration = d
}

func (t *TestTicker) getLastTickedAt() time.Time {
	t.lock.RLock()
	defer t.lock.RUnlock()
	return t.lastTickedAt
}

func (t *TestTicker) updateLastTickedAt() time.Time {
	return t.updateLastTickedAtTo(Now())
}

func (t *TestTicker) updateLastTickedAtTo(at time.Time) time.Time {
	t.lock.RLock()
	defer t.lock.RUnlock()
	t.lastTickedAt = at
	return t.lastTickedAt
}

func Since(start time.Time) time.Duration {
	return Now().Sub(start)
}

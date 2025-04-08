package internal

import (
	"time"
)

var chrono struct{ Timeline Timeline }

func init() { chrono.Timeline.Speed = 1 }

type Timeline struct {
	Altered bool
	SetAt   time.Time
	When    time.Time
	Prev    time.Time
	Frozen  bool
	Deep    bool
	Speed   float64
}

func (tl Timeline) IsZero() bool {
	return tl == Timeline{}
}

func SetSpeed(s float64) func() {
	defer notify()
	defer lock()()
	frozen := chrono.Timeline.Frozen
	td := setTime(getTime(), Option{Freeze: frozen})
	og := chrono.Timeline.Speed
	chrono.Timeline.Speed = s
	return func() {
		defer notify()
		defer lock()()
		chrono.Timeline.Speed = og
		td()
	}
}

type Option struct {
	Freeze   bool
	Unfreeze bool
	Deep     bool
}

func SetTime(target time.Time, opt Option) func() {
	defer notify()
	defer lock()()
	td := setTime(target, opt)
	return func() {
		defer notify()
		defer lock()()
		td()
	}
}

func setTime(target time.Time, opt Option) func() {
	prev := getTime()
	og := chrono.Timeline
	n := chrono.Timeline
	n.Altered = true
	n.SetAt = time.Now()
	n.Prev = prev
	n.When = target
	if opt.Freeze {
		n.Frozen = true
	}
	if opt.Deep {
		n.Deep = true
	}
	if opt.Unfreeze {
		n.Frozen = false
		n.Deep = false
	}
	chrono.Timeline = n
	return func() { chrono.Timeline = og }
}

func ScaledDuration(d time.Duration) time.Duration {
	// for some reason, two read lock at the same time has sometimes a deadlock that is not detecable with the -race conditiona detector
	// so don't use this inside other functions which are protected by rlock
	defer rlock()()
	return scaledDuration(d)
}

func scaledDuration(d time.Duration) time.Duration {
	if !chrono.Timeline.Altered {
		return d
	}
	return time.Duration(float64(d) / chrono.Timeline.Speed)
}

func RemainingDuration(from time.Time, nonScaledDuration time.Duration) time.Duration {
	defer rlock()()
	now := getTime()
	if now.Before(from) { // time travelling can be a bit weird, let's not wait forever if we went back in time
		return 0
	}
	delta := now.Sub(from)
	remainer := scaledDuration(nonScaledDuration) - delta
	if remainer < 0 { // if due to the time shift, the it was already expected
		return 0
	}
	return remainer
}

func Now() time.Time {
	defer rlock()()
	return getTime().Local()
}

func getTime() time.Time {
	now := time.Now()
	if !chrono.Timeline.Altered {
		return now
	}
	setAt := chrono.Timeline.SetAt
	if chrono.Timeline.Frozen {
		setAt = now
	}
	delta := now.Sub(setAt)
	delta = time.Duration(float64(delta) * chrono.Timeline.Speed)
	return chrono.Timeline.When.Add(delta)
}

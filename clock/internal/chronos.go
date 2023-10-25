package internal

import (
	"time"
)

func init() {
	chrono.Speed = 1
}

var chrono struct {
	Timeline struct {
		Altered bool
		SetAt   time.Time
		When    time.Time
		Frozen  bool
	}
	Speed float64
}

func SetSpeed(s float64) func() {
	defer notify()
	defer lock()()
	frozen := chrono.Timeline.Frozen
	td := setTime(getTime(), Option{Freeze: frozen})
	og := chrono.Speed
	chrono.Speed = s
	return func() {
		defer notify()
		defer lock()()
		chrono.Speed = og
		td()
	}
}

type Option struct {
	Freeze   bool
	Unfreeze bool
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
	og := chrono.Timeline
	n := chrono.Timeline
	n.Altered = true
	n.SetAt = time.Now()
	n.When = target
	if opt.Freeze {
		n.Frozen = true
	}
	if opt.Unfreeze {
		n.Frozen = false
	}
	chrono.Timeline = n
	return func() { chrono.Timeline = og }
}

func RemainingDuration(from time.Time, d time.Duration) time.Duration {
	defer rlock()()
	now := getTime()
	if now.Before(from) { // time travelling can be a bit weird, let's not wait forever if we went back in time
		return 0
	}
	scaled := time.Duration(float64(d) / chrono.Speed)
	delta := now.Sub(from)
	return scaled - delta
}

func GetTime() time.Time {
	defer rlock()()
	return getTime()
}

func getTime() time.Time {
	now := time.Now()
	if !chrono.Timeline.Altered {
		return now
	}
	if chrono.Timeline.Frozen {
		chrono.Timeline.SetAt = now
	}
	delta := now.Sub(chrono.Timeline.SetAt)
	delta = time.Duration(float64(delta) * chrono.Speed)
	return chrono.Timeline.When.Add(delta)
}

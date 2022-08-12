package internal

import (
	"github.com/adamluzsi/testcase/random"
	"time"
)

var rnd = random.New(random.CryptoSeed{})

func init() {
	chrono.Speed = 1
}

var chrono struct {
	Timeline struct {
		Altered bool
		SetAt   time.Time
		When    time.Time
	}
	Speed float64
}

func SetSpeed(s float64) func() {
	defer notify()
	defer lock()()
	setTime(time.Now())
	og := chrono.Speed
	chrono.Speed = s
	return func() {
		defer notify()
		defer lock()()
		chrono.Speed = og
	}
}

func SetTime(target time.Time) func() {
	defer notify()
	defer lock()()
	return setTime(target)
}

func setTime(target time.Time) func() {
	og := chrono.Timeline
	n := chrono.Timeline
	n.Altered = true
	n.SetAt = time.Now()
	n.When = target
	chrono.Timeline = n
	return func() {
		defer notify()
		defer lock()()
		chrono.Timeline = og
	}
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
	delta := now.Sub(chrono.Timeline.SetAt)
	delta = time.Duration(float64(delta) * chrono.Speed)
	return chrono.Timeline.When.Add(delta)
}

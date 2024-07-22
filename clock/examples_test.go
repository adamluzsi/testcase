package clock_test

import (
	"testing"
	"time"

	"go.llib.dev/testcase/assert"
	"go.llib.dev/testcase/clock"
	"go.llib.dev/testcase/clock/timecop"
)

func ExampleTimeNow_freeze() {
	var tb testing.TB

	type Entity struct {
		CreatedAt time.Time
	}

	MyFunc := func() Entity {
		return Entity{
			CreatedAt: clock.Now(),
		}
	}

	expected := Entity{
		CreatedAt: clock.Now(),
	}

	timecop.Travel(tb, expected.CreatedAt, timecop.Freeze)

	assert.Equal(tb, expected, MyFunc())
}

func ExampleTimeNow_withTravelByDuration() {
	var tb testing.TB

	_ = clock.Now() // now
	timecop.Travel(tb, time.Hour)
	_ = clock.Now() // now + 1 hour
}

func ExampleTimeNow_withTravelByDate() {
	var tb testing.TB

	date := time.Date(2022, 01, 01, 12, 0, 0, 0, time.Local)
	timecop.Travel(tb, date, timecop.Freeze) // freeze the time until it is read
	time.Sleep(time.Second)
	_ = clock.Now() // equals with date
}

func ExampleAfter() {
	var tb testing.TB
	timecop.SetSpeed(tb, 5)    // 5x time speed
	<-clock.After(time.Second) // but only wait 1/5 of the time
}

func ExampleSleep() {
	var tb testing.TB
	timecop.SetSpeed(tb, 5)  // 5x time speed
	clock.Sleep(time.Second) // but only sleeps 1/5 of the time
}

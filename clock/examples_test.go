package clock_test

import (
	"github.com/adamluzsi/testcase/clock"
	"github.com/adamluzsi/testcase/clock/timecop"
	"testing"
	"time"
)

func ExampleTimeNow() {
	_ = clock.TimeNow() // now

	var tb testing.TB
	timecop.Travel(tb, time.Hour)

	_ = clock.TimeNow() // now + 1 hour

	timecop.TravelTo(tb, 2022, 01, 01)
	_ = clock.TimeNow() // 2022-01-01 at {now.Hour}-{now.Minute}-{now.Second}
}

func ExampleAfter() {
	var tb testing.TB
	timecop.SetFlowOfTime(tb, 5) // 5x time speed
	<-clock.After(time.Second)   // but only wait 1/5 of the time
}

func ExampleSleep() {
	var tb testing.TB
	timecop.SetFlowOfTime(tb, 5) // 5x time speed
	clock.Sleep(time.Second)     // but only sleeps 1/5 of the time
}

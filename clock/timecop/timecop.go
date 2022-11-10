package timecop

import (
	"testing"
	"time"

	"github.com/adamluzsi/testcase/clock/internal"
)

func Travel[D time.Duration | time.Time](tb testing.TB, d D, tos ...TravelOption) {
	tb.Helper()
	guardAgainstParallel(tb)
	opt := toOption(tos)
	switch d := any(d).(type) {
	case time.Duration:
		travelByDuration(tb, d, opt.Freeze)
	case time.Time:
		travelByTime(tb, d, opt.Freeze)
	}
}

func SetSpeed(tb testing.TB, multiplier float64) {
	tb.Helper()
	guardAgainstParallel(tb)
	if multiplier <= 0 {
		tb.Fatal("Timecop.SetSpeed can't receive zero or negative value")
	}
	tb.Cleanup(internal.SetSpeed(multiplier))
}

// guardAgainstParallel
// is a hack that ensures that there was no testing.T.Parallel() used in the test.
func guardAgainstParallel(tb testing.TB) {
	tb.Helper()
	const key, value = `TEST_CASE_TIMECOP_IN_USE`, "TRUE"
	tb.Setenv(key, value)
}

func travelByDuration(tb testing.TB, d time.Duration, freeze bool) {
	tb.Helper()
	travelByTime(tb, internal.GetTime().Add(d), freeze)
}

func travelByTime(tb testing.TB, target time.Time, freeze bool) {
	tb.Helper()
	tb.Cleanup(internal.SetTime(target, freeze))
}

// Freeze instruct travel to freeze the time until the first time reading on the clock.
func Freeze() TravelOption {
	return fnTravelOption(func(o *option) {
		o.Freeze = true
	})
}

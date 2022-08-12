package timecop

import (
	"github.com/adamluzsi/testcase/clock/internal"
	"testing"
	"time"
)

func Travel[D time.Duration | time.Time](tb testing.TB, d D) {
	tb.Helper()
	guardAgainstParallel(tb)
	switch d := any(d).(type) {
	case time.Duration:
		travelByDuration(tb, d)
	case time.Time:
		travelByTime(tb, d)
	}
}

func TravelTo[M int | time.Month](tb testing.TB, year int, month M, day int) {
	tb.Helper()
	guardAgainstParallel(tb)
	now := internal.GetTime()
	tb.Cleanup(internal.SetTime(time.Date(
		year,
		time.Month(month),
		day,
		now.Hour(),
		now.Minute(),
		now.Second(),
		now.Nanosecond(),
		now.Location(),
	)))
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
	const key = `TEST_CASE_CLOC_IN_USE`
	tb.Helper()
	tb.Setenv(key, "TRUE")
}

func travelByDuration(tb testing.TB, d time.Duration) {
	tb.Helper()
	travelByTime(tb, internal.GetTime().Add(d))
}

func travelByTime(tb testing.TB, target time.Time) {
	tb.Helper()
	tb.Cleanup(internal.SetTime(target))
}

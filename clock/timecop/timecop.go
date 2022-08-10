package timecop

import (
	"github.com/adamluzsi/testcase/clock"
	"github.com/adamluzsi/testcase/clock/internal"
	"testing"
	"time"
)

func Travel(tb testing.TB, d time.Duration) {
	tb.Helper()
	guardAgainstParallel(tb)
	internal.Chronos.Offset += d
	tb.Cleanup(func() {
		internal.Chronos.Offset -= d
	})
}

func TravelTo[M int | time.Month](tb testing.TB, years int, months M, days int) {
	tb.Helper()
	now := clock.TimeNow()
	var (
		yd = years - now.Year()
		md = time.Month(months) - now.Month()
		dd = days - now.Day()
	)
	Travel(tb, now.AddDate(yd, int(md), dd).Sub(now))
}

func SetFlowOfTime(tb testing.TB, multiplier float64) {
	tb.Helper()
	guardAgainstParallel(tb)
	if multiplier <= 0 {
		tb.Fatal("Timecop.SetFlowOfTime can't receive zero or negative value")
	}
	og := internal.Chronos.FlowOfTime
	internal.Chronos.FlowOfTime = multiplier
	tb.Cleanup(func() { internal.Chronos.FlowOfTime = og })
}

// guardAgainstParallel ensures that there was no .Parallel() use in the test.
func guardAgainstParallel(tb testing.TB) {
	const key = `TEST_CASE_CLOC_IN_USE`
	tb.Helper()
	tb.Setenv(key, "TRUE")
}

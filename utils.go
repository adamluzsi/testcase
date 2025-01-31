package testcase

import (
	"testing"
	"time"

	"go.llib.dev/testcase/internal/env"
)

// SkipUntil is equivalent to SkipNow if the test is executing prior to the given deadline time.
// SkipUntil is useful when you need to skip something temporarily, but you don't trust your memory enough to return to it on your own.
func SkipUntil(tb testing.TB, year int, month time.Month, day int, hour int) {
	tb.Helper()
	const skipTimeFormat = "2006-01-02"
	target := time.Date(year, month, day, hour, 0, 0, 0, time.Local)
	fdate := target.Format(skipTimeFormat)
	if time.Now().Before(target) {
		tb.Skipf("Skip time %s", fdate)
	}
	tb.Logf("[SkipUntil] expired on %s", fdate)
	tb.Log("consider removing [SkipUntil]")
}

// OnFail will execute a funcion block in case the test fails.
func OnFail(tb testing.TB, fn func()) {
	tb.Helper()
	tb.Cleanup(func() {
		tb.Helper()
		if tb.Failed() {
			fn()
		}
	})
}

//-------------------------------------------------- Env Var Helpers -------------------------------------------------//

// SetEnv will set the os environment variable for the current program to a given value,
// and prepares a cleanup function to restore the original state of the environment variable.
//
// Spec using this helper should be flagged with Spec.HasSideEffect or Spec.Sequential.
func SetEnv(tb testing.TB, key, value string) {
	tb.Helper()
	tb.Setenv(key, value)
	env.SetEnv(tb, key, value)
}

// UnsetEnv will unset the os environment variable value for the current program,
// and prepares a cleanup function to restore the original state of the environment variable.
//
// Spec using this helper should be flagged with Spec.HasSideEffect or Spec.Sequential.
func UnsetEnv(tb testing.TB, key string) {
	tb.Helper()
	//tb.Setenv(key, "") // to trigger parallel error check
	env.UnsetEnv(tb, key)
}

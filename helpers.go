package testcase

import (
	"os"
	"testing"
)

// SetEnv will set the os environment variable for the current program to a given value,
// and prepares a cleanup function to restore the initial state of the environment variable.
//
// Spec using this helper should be flagged with Spec.HasSideEffect or Spec.Sequential.
func SetEnv(tb testing.TB, key, value string) {
	var restore func() error
	if originalValue, ok := os.LookupEnv(key); ok {
		restore = func() error { return os.Setenv(key, originalValue) }
	} else {
		restore = func() error { return os.Unsetenv(key) }
	}
	tb.Cleanup(func() {
		if err := restore(); err != nil {
			tb.Error(err)
		}
	})
	if err := os.Setenv(key, value); err != nil {
		tb.Fatal(err)
	}
}

package env

import "os"

type TB interface {
	Helper()
	Cleanup(func())
	Fatal(...any)
	Error(...any)
}

// SetEnv will set the os environment variable for the current program to a given value,
// and prepares a cleanup function to restore the original state of the environment variable.
//
// Spec using this helper should be flagged with Spec.HasSideEffect or Spec.Sequential.
func SetEnv(tb TB, key, value string) {
	tb.Helper()
	cleanupEnv(tb, key)

	if err := os.Setenv(key, value); err != nil {
		tb.Fatal(err)
	}
}

// UnsetEnv will unset the os environment variable value for the current program,
// and prepares a cleanup function to restore the original state of the environment variable.
//
// Spec using this helper should be flagged with Spec.HasSideEffect or Spec.Sequential.
func UnsetEnv(tb TB, key string) {
	tb.Helper()
	cleanupEnv(tb, key)

	if err := os.Unsetenv(key); err != nil {
		tb.Fatal(err)
	}
}

func cleanupEnv(tb TB, key string) {
	tb.Helper()
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
}

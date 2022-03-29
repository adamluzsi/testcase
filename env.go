package testcase

import (
	"os"
	"testing"
)

// EnvKeySeed is the environment variable key that will be checked for a pseudo random seed,
// which will be used to randomize the order of executions between test cases.
const EnvKeySeed = `TESTCASE_SEED`

// EnvKeyOrdering is the environment variable key that will be checked for testCase determine
// what order of execution should be used between test cases in a testing group.
// The default sorting behavior is pseudo random based on an the seed.
//
// Mods:
// - defined: execute testCase in the order which they are being defined
// - random: pseudo random based ordering between tests.
const EnvKeyOrdering = `TESTCASE_ORDERING`

//-------------------------------------------------- Env Var Helpers -------------------------------------------------//

// SetEnv will set the os environment variable for the current program to a given value,
// and prepares a cleanup function to restore the original state of the environment variable.
//
// Spec using this helper should be flagged with Spec.HasSideEffect or Spec.Sequential.
func SetEnv(tb testing.TB, key, value string) {
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
func UnsetEnv(tb testing.TB, key string) {
	tb.Helper()
	cleanupEnv(tb, key)

	if err := os.Unsetenv(key); err != nil {
		tb.Fatal(err)
	}
}

func cleanupEnv(tb testing.TB, key string) {
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

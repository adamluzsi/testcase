package assert

import "testing"

func MakeIt(tb testing.TB) It {
	return It{
		TB:     tb,
		Must:   Must(tb),
		Should: Should(tb),
	}
}

type It struct {
	testing.TB
	// Must Asserter will use FailNow on a failed assertion.
	// This will make test exit early on.
	Must Asserter
	// Should Asserter's will allow to continue the test scenario,
	// but mark test failed on a failed assertion.
	Should Asserter
}

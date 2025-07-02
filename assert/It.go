package assert

import "testing"

// MakeIt will make an It.
//
// Deprecated: use assert top level functions directly, like Must and Should
func MakeIt(tb testing.TB) It {
	return It{
		TB:     tb,
		Must:   Must(tb),
		Should: Should(tb),
	}
}

// It
//
// Deprecated: assert package functions instead.
type It struct {
	testing.TB
	// Must Asserter will use FailNow on a failed assertion.
	// This will make test exit early on.
	Must Asserter
	// Should Asserter's will allow to continue the test scenario,
	// but mark test failed on a failed assertion.
	Should Asserter
}

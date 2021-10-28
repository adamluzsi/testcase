package assert

import "testing"

func makeIt(tb testing.TB) It {
	return It{
		Must:   Must(tb),
		Should: Should(tb),
	}
}

type It struct {
	Must   Asserter
	Should Asserter
}

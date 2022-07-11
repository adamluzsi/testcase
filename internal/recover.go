package internal

import (
	"github.com/adamluzsi/testcase/sandbox"
)

// RecoverGoexit helps overcome the testing.TB#FailNow's behaviour
// where on failure the goroutine exits to finish earlier.
func RecoverGoexit(fn func()) sandbox.RunOutcome {
	runOutcome := sandbox.Run(fn)
	if runOutcome.Goexit { // ignore goexit
		return runOutcome
	}
	if !runOutcome.OK { // propagate panic
		panic(runOutcome.PanicValue)
	}
	return runOutcome
}

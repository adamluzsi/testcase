package testcase_test

import (
	"github.com/adamluzsi/testcase"
)

type ExampleRaceSafe struct{}

func (ExampleRaceSafe) ThreadSafeCall() {}

func ExampleRace() {
	v := ExampleRaceSafe{}

	// running `go test` with the `-race` flag should help you detect unsafe implementations.
	testcase.Race(func() {
		// this will run in multiple instance, with race.
		v.ThreadSafeCall()
	})
}

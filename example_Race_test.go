package testcase_test

import (
	"github.com/adamluzsi/testcase"
)

type ExampleRaceSafe struct{}

func (ExampleRaceSafe) ThreadSafeCall() {}

func ExampleRace() {
	v := ExampleRaceSafe{}

	// running `go test` with the `-race` flag should help you detect unsafe implementations.
	// each block run at the same time in a race situation
	testcase.Race(func() {
		v.ThreadSafeCall()
	}, func() {
		v.ThreadSafeCall()
	})
}

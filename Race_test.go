// +build !race

package testcase_test

// The build tag "race" is defined when building with the -race flag.

import (
	"fmt"
	"testing"
	"time"

	"github.com/adamluzsi/testcase"
	"github.com/stretchr/testify/require"
)

func TestRace(t *testing.T) {
	retry := testcase.Retry{Strategy: testcase.Waiter{WaitTimeout: time.Second}}

	retry.Assert(t, func(tb testing.TB) {
		var counter int
		participants := testcase.Race(func() {
			c := counter
			time.Sleep(time.Millisecond)
			counter = c + 1 // counter++ would not work
		})

		tb.Log(`counter`, counter, `number of participants`, participants)
		require.True(t, counter < participants,
			fmt.Sprintf(`counter was expected to be less that the total number of race patricipants because we are in a race condition`))
	})
}

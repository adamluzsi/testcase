package testcase_test

import (
	"os"
	"testing"
)

func Spike(tb testing.TB) {
	if _, ok := os.LookupEnv(`SPIKE`); !ok {
		tb.Skip()
	}
}

// This spike meant to help manually verify the grouping mechanism of *testing.T
// > SPIKE=TRUE go testCase -run TestRunGroup -v
//
func TestRunGroup(t *testing.T) {
	Spike(t)

	t.Run(`single`, func(t *testing.T) {
		t.Run(`foo`, func(t *testing.T) {
			t.Run(`bar`, func(t *testing.T) {
				t.Log(`foo-bar`)
			})
			t.Run(`baz`, func(t *testing.T) {
				t.Log(`foo-baz`)
			})
		})
	})
	t.Run(`split`, func(t *testing.T) {
		t.Run(`foo`, func(t *testing.T) {
			t.Run(`bar`, func(t *testing.T) {
				t.Log(`foo-bar`)
			})
		})
		t.Run(`foo`, func(t *testing.T) {
			t.Run(`baz`, func(t *testing.T) {
				t.Log(`foo-baz`)
			})
		})
	})
	t.Run(`single+split with lambda`, func(t *testing.T) {
		var eventually []func()
		t.Run(`foo`, func(t *testing.T) {
			eventually = append(eventually, func() {
				t.Run(`bar`, func(t *testing.T) {
					t.Log(`foo-bar`)
				})
			})
			eventually = append(eventually, func() {
				t.Run(`baz`, func(t *testing.T) {
					t.Log(`foo-baz`)
				})
			})

			// this will run list of them, because we are still within the `testing#T.Run` scope
			for _, e := range eventually {
				e()
			}
		})

		// this will not run the testCase since foo `testing#T.Run` scope is closed
		for _, e := range eventually {
			e()
		}

	})
}

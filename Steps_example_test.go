package testcase_test

import (
	"testing"

	"github.com/adamluzsi/testcase"
	"github.com/stretchr/testify/require"
)

func ExampleSteps_AddWithTeardown(t *testing.T) {
	var value string

	var steps = testcase.Steps{}
	t.Run(`on`, func(t *testing.T) {
		steps := steps.AddWithTeardown(func(t *testing.T) func() { value = "1"; return func() {} })

		t.Run(`each`, func(t *testing.T) {
			steps := steps.AddWithTeardown(func(t *testing.T) func() { value = "2"; return func() {} })

			t.Run(`nested`, func(t *testing.T) {
				steps := steps.AddWithTeardown(func(t *testing.T) func() { value = "3"; return func() {} })

				t.Run(`layer`, func(t *testing.T) {
					steps := steps.AddWithTeardown(func(t *testing.T) func() { value = "4"; return func() {} })

					t.Run(`it will setup and break down the right context`, func(t *testing.T) {
						defer steps.Setup(t)()

						require.Equal(t, "4", value)
					})
				})

				t.Run(`then`, func(t *testing.T) {
					defer steps.Setup(t)()

					require.Equal(t, "3", value)
				})
			})

			t.Run(`then`, func(t *testing.T) {
				defer steps.Setup(t)()

				require.Equal(t, "2", value)
			})
		})

		t.Run(`then`, func(t *testing.T) {
			defer steps.Setup(t)()

			require.Equal(t, "1", value)
		})
	})
}

func ExampleSteps_Add(t *testing.T) {
	var value string

	var steps = testcase.Steps{}

	t.Run(`on`, func(t *testing.T) {
		steps := steps.Add(func(t *testing.T) { value = "1" })

		t.Run(`each`, func(t *testing.T) {
			steps := steps.Add(func(t *testing.T) { value = "2" })

			t.Run(`nested`, func(t *testing.T) {
				steps := steps.Add(func(t *testing.T) { value = "3" })

				t.Run(`layer`, func(t *testing.T) {
					steps := steps.Add(func(t *testing.T) { value = "4" })

					t.Run(`it will setup and break down the right context`, func(t *testing.T) {
						defer steps.Setup(t)()

						require.Equal(t, "4", value)
					})
				})

				t.Run(`then`, func(t *testing.T) {
					defer steps.Setup(t)()

					require.Equal(t, "3", value)
				})
			})

			t.Run(`then`, func(t *testing.T) {
				defer steps.Setup(t)()

				require.Equal(t, "2", value)
			})
		})

		t.Run(`then`, func(t *testing.T) {
			defer steps.Setup(t)()

			require.Equal(t, "1", value)
		})
	})
}

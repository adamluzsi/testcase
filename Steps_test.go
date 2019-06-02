package testcase_test

import (
	"testing"

	"github.com/adamluzsi/testcase"
	"github.com/stretchr/testify/require"
)

func TestSteps_AddWithTeardown(t *testing.T) {
	var value string
	var teardowns []int

	var steps = testcase.Steps{}.Around(func(t *testing.T) func() {
		teardowns = nil
		return func() {}
	})

	t.Run(`on`, func(t *testing.T) {
		steps := steps.Around(func(t *testing.T) func() {
			value = "1"
			return func() { teardowns = append(teardowns, 1) }
		})

		t.Run(`each`, func(t *testing.T) {
			steps := steps.Around(func(t *testing.T) func() {
				value = "2"
				return func() { teardowns = append(teardowns, 2) }
			})

			t.Run(`nested`, func(t *testing.T) {
				steps := steps.Around(func(t *testing.T) func() {
					value = "3"
					return func() { teardowns = append(teardowns, 3) }
				})

				t.Run(`layer`, func(t *testing.T) {
					steps := steps.Around(func(t *testing.T) func() {
						value = "4"
						return func() { teardowns = append(teardowns, 4) }
					})

					t.Run(`it will setup and break down the right context`, func(t *testing.T) {
						td := steps.Setup(t)

						require.Equal(t, "4", value)

						require.NotEqual(t, []int{1, 2, 3, 4}, teardowns)
						td()
						require.Equal(t, []int{1, 2, 3, 4}, teardowns)
					})
				})

				t.Run(`then`, func(t *testing.T) {
					td := steps.Setup(t)

					require.Equal(t, "3", value)

					require.NotEqual(t, []int{1, 2, 3}, teardowns)
					td()
					require.Equal(t, []int{1, 2, 3}, teardowns)
				})
			})

			t.Run(`then`, func(t *testing.T) {
				td := steps.Setup(t)

				require.Equal(t, "2", value)

				require.NotEqual(t, []int{1, 2}, teardowns)
				td()
				require.Equal(t, []int{1, 2}, teardowns)
			})
		})

		t.Run(`then`, func(t *testing.T) {
			td := steps.Setup(t)

			require.Equal(t, "1", value)

			require.NotEqual(t, []int{1}, teardowns)
			td()
			require.Equal(t, []int{1}, teardowns)
		})
	})
}

func TestSteps_Add(t *testing.T) {
	var value string

	var steps = testcase.Steps{}

	t.Run(`on`, func(t *testing.T) {
		steps := steps.Before(func(t *testing.T) { value = "1" })

		t.Run(`each`, func(t *testing.T) {
			steps := steps.Before(func(t *testing.T) { value = "2" })

			t.Run(`nested`, func(t *testing.T) {
				steps := steps.Before(func(t *testing.T) { value = "3" })

				t.Run(`layer`, func(t *testing.T) {
					steps := steps.Before(func(t *testing.T) { value = "4" })

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

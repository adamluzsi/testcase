package testrunctx_test

import (
	"github.com/adamluzsi/testrunctx"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestSetup(t *testing.T) {
	var value string

	var steps []func(*testing.T)

	t.Run(`on`, func(t *testing.T) {
		steps := append(steps, func(t *testing.T) { value = "1" })

		t.Run(`each`, func(t *testing.T) {
			steps := append(steps, func(t *testing.T) { value = "2" })

			t.Run(`nested`, func(t *testing.T) {
				steps := append(steps, func(t *testing.T) { value = "3" })

				t.Run(`layer`, func(t *testing.T) {
					steps := append(steps, func(t *testing.T) { value = "4" })

					t.Run(`it will setup and break down the right context`, func(t *testing.T) {
						testrunctx.Setup(t, steps...)

						require.Equal(t, "4", value)
					})
				})

				t.Run(`then`, func(t *testing.T) {
					testrunctx.Setup(t, steps...)

					require.Equal(t, "3", value)
				})
			})

			t.Run(`then`, func(t *testing.T) {
				testrunctx.Setup(t, steps...)

				require.Equal(t, "2", value)
			})
		})

		t.Run(`then`, func(t *testing.T) {
			testrunctx.Setup(t, steps...)

			require.Equal(t, "1", value)
		})
	})
}

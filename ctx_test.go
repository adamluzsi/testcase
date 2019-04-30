package testrun_test

import (
	"github.com/adamluzsi/testrun"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestCTX(t *testing.T) {
	subject := func() *testrun.CTX { return testrun.NewCTX() }

	t.Run(`Step`, func(t *testing.T) {
		testCTX_Step(t, subject())
	})

	t.Run(`Let`, func(t *testing.T) {
		testCTX_Let(t, subject())
	})
}

func testCTX_Step(t *testing.T, ctx *testrun.CTX) {
	var value string

	defer ctx.Step(func(t *testing.T) {
		value = ""
	})()

	t.Run(`on`, func(t *testing.T) {
		defer ctx.Step(func(t *testing.T) { value = "1" })()

		t.Run(`each`, func(t *testing.T) {
			defer ctx.Step(func(t *testing.T) { value = "2" })()

			t.Run(`nested`, func(t *testing.T) {
				defer ctx.Step(func(t *testing.T) { value = "3" })()

				t.Run(`layer`, func(t *testing.T) {
					defer ctx.Step(func(t *testing.T) { value = "4" })()

					t.Run(`it will setup and break down the right context`, func(t *testing.T) {
						ctx.Setup(t)

						require.Equal(t, "4", value)
					})
				})

				t.Run(`then`, func(t *testing.T) {
					ctx.Setup(t)

					require.Equal(t, "3", value)
				})
			})

			t.Run(`then`, func(t *testing.T) {
				ctx.Setup(t)

				require.Equal(t, "2", value)
			})
		})

		t.Run(`then`, func(t *testing.T) {
			ctx.Setup(t)

			require.Equal(t, "1", value)
		})
	})
}

func testCTX_Let(t *testing.T, ctx *testrun.CTX) {

	defer ctx.Let(`x`, func(vars testrun.Vars) interface{} { return "" })()

	defer ctx.Let(`y`, func(vars testrun.Vars) interface{} {
		var x string
		vars.Get("x", &x)
		return x
	})()

	t.Run(`on`, func(t *testing.T) {
		defer ctx.Let(`x`, func(vars testrun.Vars) interface{} { return "1" })()

		t.Run(`each`, func(t *testing.T) {
			defer ctx.Let(`x`, func(vars testrun.Vars) interface{} { return "2" })()

			t.Run(`nested`, func(t *testing.T) {
				defer ctx.Let(`x`, func(vars testrun.Vars) interface{} { return "3" })()

				t.Run(`layer`, func(t *testing.T) {
					defer ctx.Let(`x`, func(vars testrun.Vars) interface{} { return "4" })()

					t.Run(`it will setup and break down the right context`, func(t *testing.T) {
						vars := ctx.Setup(t)
						t.Parallel()

						var value string
						vars.Get("x", &value)
						require.Equal(t, "4", value)
					})
				})

				t.Run(`when one var ref to another`, func(t *testing.T) {
					// the y value above

					t.Run(`then it should be able to get the referenced value on the latest version`, func(t *testing.T) {
						vars := ctx.Setup(t)
						t.Parallel()

						var value string
						vars.Get("y", &value)
						require.Equal(t, "3", value)
					})
				})

				t.Run(`then`, func(t *testing.T) {
					vars := ctx.Setup(t)
					t.Parallel()

					var value string
					vars.Get("x", &value)
					require.Equal(t, "3", value)
				})
			})

			t.Run(`then`, func(t *testing.T) {
				vars := ctx.Setup(t)
				t.Parallel()
				var value string
				vars.Get("x", &value)
				require.Equal(t, "2", value)
			})
		})

		t.Run(`then`, func(t *testing.T) {
			vars := ctx.Setup(t)
			t.Parallel()
			var value string
			vars.Get("x", &value)
			require.Equal(t, "1", value)
		})
	})
}

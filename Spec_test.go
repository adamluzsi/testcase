package testcase_test

import (
	"github.com/adamluzsi/testcase"
	"github.com/stretchr/testify/require"
	"math/rand"
	"strconv"
	"testing"
)

func TestSpec(t *testing.T) {
	ExampleNewSpec(t)
}

func TestSmokeTest(t *testing.T) {
	spec := testcase.NewSpec(t)

	var sideEffect []string
	var currentSE []string

	valueName := strconv.Itoa(rand.Int())
	nest1Value := rand.Int()
	nest2Value := rand.Int()
	nest3Value := rand.Int()

	// I know this is cheating
	spec.Before(func(t *testing.T) {
		currentSE = make([]string, 0)
	})

	spec.Describe(`nest-lvl-1`, func(t *testing.T) {
		subject := func(v *testcase.V) int { return v.I(valueName).(int) }
		spec.Let(valueName, func(v *testcase.V) interface{} { return nest1Value })

		spec.When(`nest-lvl-2`, func(t *testing.T) {
			spec.Let(valueName, func(v *testcase.V) interface{} { return nest2Value })

			spec.After(func(t *testing.T) {
				sideEffect = append(sideEffect, "after1")
			})

			spec.Before(func(t *testing.T) {
				currentSE = append(currentSE, `before1`)
				sideEffect = append(sideEffect, `before1`)
			})

			spec.Around(func(t *testing.T) func() {
				currentSE = append(currentSE, `around1-begin`)
				sideEffect = append(sideEffect, `around1-begin`)
				return func() {
					sideEffect = append(sideEffect, `around1-end`)
				}
			})

			spec.And(`nest-lvl-3`, func(t *testing.T) {
				spec.Let(valueName, func(v *testcase.V) interface{} { return nest3Value })

				spec.After(func(t *testing.T) {
					sideEffect = append(sideEffect, "after2")
				})

				spec.Before(func(t *testing.T) {
					currentSE = append(currentSE, `before2`)
					sideEffect = append(sideEffect, `before2`)
				})

				spec.Around(func(t *testing.T) func() {
					currentSE = append(currentSE, `around2-begin`)
					sideEffect = append(sideEffect, `around2-begin`)
					return func() {
						sideEffect = append(sideEffect, `around2-end`)
					}
				})

				spec.Then(`lvl-3`, func(t *testing.T, v *testcase.V) {
					expectedCurrentSE := []string{`before1`, `around1-begin`, `before2`, `around2-begin`}
					require.Equal(t, expectedCurrentSE, currentSE)
					// t.Parallel()

					require.Equal(t, nest3Value, v.I(valueName))
					require.Equal(t, nest3Value, subject(v))
				})
			})

			spec.Then(`lvl-2`, func(t *testing.T, v *testcase.V) {
				require.Equal(t, []string{`before1`, `around1-begin`}, currentSE)
				// t.Parallel()

				require.Equal(t, nest2Value, v.I(valueName))
				require.Equal(t, nest2Value, subject(v))
			})
		})

		spec.Then(`lvl-1`, func(t *testing.T, v *testcase.V) {
			require.Equal(t, []string{}, currentSE)
			// t.Parallel()

			require.Equal(t, nest1Value, v.I(valueName))
			require.Equal(t, nest1Value, subject(v))
		})
	})

	expectedAllSideEffects := []string{
		// nest-lvl-2
		"before1", "around1-begin", "before2", "around2-begin",
		"after1", "around1-end", "after2", "around2-end",

		// nest-lvl-1
		"before1", "around1-begin",
		"after1", "around1-end",
	}

	require.Equal(t, expectedAllSideEffects, sideEffect)

}

func TestParallel(t *testing.T) {
	spec := testcase.NewSpec(t)

	valueName := strconv.Itoa(rand.Int())
	nest1Value := rand.Int()
	nest2Value := rand.Int()
	nest3Value := rand.Int()

	spec.Describe(`nest-lvl-1`, func(t *testing.T) {
		subject := func(v *testcase.V) int { return v.I(valueName).(int) }
		spec.Let(valueName, func(v *testcase.V) interface{} { return nest1Value })

		spec.When(`nest-lvl-2`, func(t *testing.T) {
			spec.Let(valueName, func(v *testcase.V) interface{} { return nest2Value })

			spec.And(`nest-lvl-3`, func(t *testing.T) {
				spec.Let(valueName, func(v *testcase.V) interface{} { return nest3Value })

				spec.Then(`lvl-3`, func(t *testing.T, v *testcase.V) {
					t.Parallel()
					require.Equal(t, nest3Value, v.I(valueName))
					require.Equal(t, nest3Value, subject(v))
				})
			})

			spec.Then(`lvl-2`, func(t *testing.T, v *testcase.V) {
				t.Parallel()
				require.Equal(t, nest2Value, v.I(valueName))
				require.Equal(t, nest2Value, subject(v))
			})
		})

		spec.Then(`lvl-1`, func(t *testing.T, v *testcase.V) {
			t.Parallel()
			require.Equal(t, nest1Value, v.I(valueName))
			require.Equal(t, nest1Value, subject(v))
		})
	})

}

func TestInvalidHookUsage(t *testing.T) {
	spec := testcase.NewSpec(t)

	valueName := strconv.Itoa(rand.Int())
	nest1Value := rand.Int()
	nest2Value := rand.Int()
	nest3Value := rand.Int()

	willPanic := func(block func()) (panicked bool) {

		defer func() {
			if r := recover(); r != nil {
				panicked = true
			}
		}()

		block()

		return false

	}

	panicSpecs := func(t *testing.T, expectedToPanic bool) {
		require.Equal(t, expectedToPanic, willPanic(func() {
			spec.Before(func(t *testing.T) {})
		}))

		require.Equal(t, expectedToPanic, willPanic(func() {
			spec.After(func(t *testing.T) {})
		}))

		require.Equal(t, expectedToPanic, willPanic(func() {
			spec.Around(func(t *testing.T) func() { return func() {} })
		}))

		require.Equal(t, expectedToPanic, willPanic(func() {
			spec.Let(strconv.Itoa(rand.Int()), func(v *testcase.V) interface{} { return nil })
		}))
	}

	shouldPanicForHooking := func(t *testing.T) { panicSpecs(t, true) }
	shouldNotPanicForHooking := func(t *testing.T) { panicSpecs(t, false) }

	spec.Describe(`nest-lvl-1`, func(t *testing.T) {

		shouldNotPanicForHooking(t)
		spec.Let(valueName, func(v *testcase.V) interface{} { return nest1Value })

		shouldNotPanicForHooking(t)
		spec.Then(`lvl-1-first`, func(t *testing.T, v *testcase.V) {})

		shouldPanicForHooking(t)

		spec.When(`nest-lvl-2`, func(t *testing.T) {
			shouldNotPanicForHooking(t)
			spec.Let(valueName, func(v *testcase.V) interface{} { return nest2Value })

			shouldNotPanicForHooking(t)
			spec.And(`nest-lvl-3`, func(t *testing.T) {
				spec.Let(valueName, func(v *testcase.V) interface{} { return nest3Value })

				spec.Then(`lvl-3`, func(t *testing.T, v *testcase.V) {})

				shouldPanicForHooking(t)

			})

			shouldPanicForHooking(t)

			spec.Then(`lvl-2`, func(t *testing.T, v *testcase.V) {})

			shouldPanicForHooking(t)

			spec.And(`nest-lvl-2-2`, func(t *testing.T) {
				shouldNotPanicForHooking(t)
				spec.Then(`nest-lvl-2-2-then`, func(t *testing.T, v *testcase.V) {})
				shouldPanicForHooking(t)
			})

			shouldPanicForHooking(t)

		})

		shouldPanicForHooking(t)

		spec.Then(`lvl-1-last`, func(t *testing.T, v *testcase.V) {})

		shouldPanicForHooking(t)

	})
}

func TestFriendlyVarNotDefined(t *testing.T) {
	spec := testcase.NewSpec(t)

	getPanicMessage := func(block func()) (msg string) {
		defer func() {
			if r := recover(); r != nil {
				msg = r.(string)
			}
		}()

		block()
		return ""
	}

	spec.Let(`var1`, func(v *testcase.V) interface{} { return `hello-world` })
	spec.Let(`var2`, func(v *testcase.V) interface{} { return `hello-world` })

	spec.Then(`var1 var found`, func(t *testing.T, v *testcase.V) {
		require.Equal(t, `hello-world`, v.I(`var1`).(string))
	})

	spec.Then(`not existing var will panic with friendly msg`, func(t *testing.T, v *testcase.V) {
		panicMSG := getPanicMessage(func() { v.I(`not-exist`) })
		require.Contains(t, panicMSG, `Variable "not-exist" is not found`)
		require.Contains(t, panicMSG, `Did you mean?`)
		require.Contains(t, panicMSG, `var1`)
		require.Contains(t, panicMSG, `var2`)
	})

}

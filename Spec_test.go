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

func TestSpec_SmokeTest(t *testing.T) {
	spec := testcase.NewSpec(t)

	var sideEffect []string
	var currentSE []string

	valueName := strconv.Itoa(rand.Int())
	nest1Value := rand.Int()
	nest2Value := rand.Int()
	nest3Value := rand.Int()

	// I know this is cheating
	spec.Before(func(t *testing.T, v *testcase.V) {
		currentSE = make([]string, 0)
	})

	spec.Describe(`nest-lvl-1`, func(spec *testcase.Spec) {
		subject := func(v *testcase.V) int { return v.I(valueName).(int) }
		spec.Let(valueName, func(v *testcase.V) interface{} { return nest1Value })

		spec.When(`nest-lvl-2`, func(spec *testcase.Spec) {
			spec.Let(valueName, func(v *testcase.V) interface{} { return nest2Value })

			spec.After(func(t *testing.T, v *testcase.V) {
				sideEffect = append(sideEffect, "after1")
			})

			spec.Before(func(t *testing.T, v *testcase.V) {
				currentSE = append(currentSE, `before1`)
				sideEffect = append(sideEffect, `before1`)
			})

			spec.Around(func(t *testing.T, v *testcase.V) func() {
				currentSE = append(currentSE, `around1-begin`)
				sideEffect = append(sideEffect, `around1-begin`)
				return func() {
					sideEffect = append(sideEffect, `around1-end`)
				}
			})

			spec.And(`nest-lvl-3`, func(spec *testcase.Spec) {
				spec.Let(valueName, func(v *testcase.V) interface{} { return nest3Value })

				spec.After(func(t *testing.T, v *testcase.V) {
					sideEffect = append(sideEffect, "after2")
				})

				spec.Before(func(t *testing.T, v *testcase.V) {
					currentSE = append(currentSE, `before2`)
					sideEffect = append(sideEffect, `before2`)
				})

				spec.Around(func(t *testing.T, v *testcase.V) func() {
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

func TestSpec_ParallelSupport(t *testing.T) {
	spec := testcase.NewSpec(t)

	valueName := strconv.Itoa(rand.Int())
	nest1Value := rand.Int()
	nest2Value := rand.Int()
	nest3Value := rand.Int()

	spec.Describe(`nest-lvl-1`, func(spec *testcase.Spec) {
		subject := func(v *testcase.V) int { return v.I(valueName).(int) }
		spec.Let(valueName, func(v *testcase.V) interface{} { return nest1Value })

		spec.When(`nest-lvl-2`, func(spec *testcase.Spec) {
			spec.Let(valueName, func(v *testcase.V) interface{} { return nest2Value })

			spec.And(`nest-lvl-3`, func(spec *testcase.Spec) {
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

func TestSpec_InvalidHookUsage(t *testing.T) {
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

	panicSpecs := func(t *testing.T, spec *testcase.Spec, expectedToPanic bool) {
		require.Equal(t, expectedToPanic, willPanic(func() {
			spec.Before(func(t *testing.T, v *testcase.V) {})
		}))

		require.Equal(t, expectedToPanic, willPanic(func() {
			spec.After(func(t *testing.T, v *testcase.V) {})
		}))

		require.Equal(t, expectedToPanic, willPanic(func() {
			spec.Around(func(t *testing.T, v *testcase.V) func() { return func() {} })
		}))

		require.Equal(t, expectedToPanic, willPanic(func() {
			spec.Let(strconv.Itoa(rand.Int()), func(v *testcase.V) interface{} { return nil })
		}))
	}

	shouldPanicForHooking := func(t *testing.T, s *testcase.Spec) { panicSpecs(t, s, true) }
	shouldNotPanicForHooking := func(t *testing.T, s *testcase.Spec) { panicSpecs(t, s, false) }

	topSpec := spec

	spec.Describe(`nest-lvl-1`, func(spec *testcase.Spec) {

		shouldPanicForHooking(t, topSpec)

		shouldNotPanicForHooking(t, spec)
		spec.Let(valueName, func(v *testcase.V) interface{} { return nest1Value })

		shouldNotPanicForHooking(t, spec)
		spec.Then(`lvl-1-first`, func(t *testing.T, v *testcase.V) {})

		shouldPanicForHooking(t, spec)

		spec.When(`nest-lvl-2`, func(spec *testcase.Spec) {
			shouldNotPanicForHooking(t, spec)
			spec.Let(valueName, func(v *testcase.V) interface{} { return nest2Value })

			shouldNotPanicForHooking(t, spec)
			spec.And(`nest-lvl-3`, func(spec *testcase.Spec) {
				spec.Let(valueName, func(v *testcase.V) interface{} { return nest3Value })

				spec.Then(`lvl-3`, func(t *testing.T, v *testcase.V) {})

				shouldPanicForHooking(t, spec)
			})

			shouldPanicForHooking(t, spec)

			spec.Then(`lvl-2`, func(t *testing.T, v *testcase.V) {})

			shouldPanicForHooking(t, spec)

			spec.And(`nest-lvl-2-2`, func(spec *testcase.Spec) {
				shouldNotPanicForHooking(t, spec)
				spec.Then(`nest-lvl-2-2-then`, func(t *testing.T, v *testcase.V) {})
				shouldPanicForHooking(t, spec)
			})

			shouldPanicForHooking(t, spec)

		})

		shouldPanicForHooking(t, spec)

		spec.Then(`lvl-1-last`, func(t *testing.T, v *testcase.V) {})

		shouldPanicForHooking(t, spec)

	})
}

func TestSpec_FriendlyVarNotDefined(t *testing.T) {
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

func TestSpec_VarValuesAreDeterministicallyCached(t *testing.T) {
	spec := testcase.NewSpec(t)

	var testCase1Value int
	var testCase2Value int

	spec.Describe(`Let`, func(s *testcase.Spec) {
		s.Let(`int`, func(v *testcase.V) interface{} { return rand.Int() })

		s.Then(`regardless of multiple call, let value remain the same for each`, func(t *testing.T, v *testcase.V) {
			value := v.I(`int`).(int)
			testCase1Value = value
			require.Equal(t, value, v.I(`int`).(int))
		})

		s.Then(`for every then block then block value is reevaluated`, func(t *testing.T, v *testcase.V) {
			value := v.I(`int`).(int)
			testCase2Value = value
			require.Equal(t, value, v.I(`int`).(int))
		})

	})

	require.NotEqual(t, testCase1Value, testCase2Value)
}

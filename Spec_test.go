package testcase_test

import (
	"math/rand"
	"strconv"
	"testing"

	"github.com/adamluzsi/testcase"
	"github.com/stretchr/testify/require"
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

func TestSpec_ParallelSafeVariableSupport(t *testing.T) {
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

func TestSpec_InvalidUsages(t *testing.T) {
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

		require.Equal(t, expectedToPanic, willPanic(func() {
			spec.Parallel()
		}))

		//require.Equal(t, expectedToPanic, willPanic(func() {
		//	spec.LetNow(`value`, int(42))
		//}))
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

		s.And(`the value is accessible from the hooks as well`, func(s *testcase.Spec) {
			var value int

			s.Before(func(t *testing.T, v *testcase.V) {
				value = v.I(`int`).(int)
			})

			s.Then(`it will remain the same value in the test case as well compared to the before block`, func(t *testing.T, v *testcase.V) {
				require.NotEqual(t, 0, value)
				require.Equal(t, value, v.I(`int`).(int))
			})
		})

		s.And(`struct value can be modified by hooks for preparation purposes like setting up mocks expectations`, func(s *testcase.Spec) {
			s.Let(`struct`, func(v *testcase.V) interface{} {
				return &MyType{}
			})

			s.Before(func(t *testing.T, v *testcase.V) {
				value := v.I(`struct`).(*MyType)
				value.Field1 = "testing"
			})

			s.Then(`the value can be seen from the test case scope`, func(t *testing.T, v *testcase.V) {
				require.Equal(t, `testing`, v.I(`struct`).(*MyType).Field1)
			})
		})
	})

	require.NotEqual(t, testCase1Value, testCase2Value)
}

func TestSpec_Parallel(t *testing.T) {
	s := testcase.NewSpec(t)

	isPanic := func(block func()) (panicked bool) {
		defer func() {
			if r := recover(); r != nil {
				panicked = true
			}
		}()
		block()
		return false
	}

	s.Describe(`Parallel`, func(s *testcase.Spec) {

		s.When(`no parallel set on top level nesting`, func(s *testcase.Spec) {
			s.And(`on each sub level`, func(s *testcase.Spec) {
				s.Then(`it will accept T#Parallel call`, func(t *testing.T, v *testcase.V) {
					require.False(t, isPanic(func() { t.Parallel() }))
				})
			})
			s.Then(`it will accept T#Parallel call`, func(t *testing.T, v *testcase.V) {
				require.False(t, isPanic(func() { t.Parallel() }))
			})
		})

		s.When(`on the first level there is no parallel configured`, func(s *testcase.Spec) {
			s.And(`on the second one, yes`, func(s *testcase.Spec) {
				s.Parallel()

				s.And(`parallel will be "inherited" for each nested context`, func(s *testcase.Spec) {
					s.Then(`it will panic on T#Parallel call`, func(t *testing.T, v *testcase.V) {
						require.True(t, isPanic(func() { t.Parallel() }))
					})
				})

				s.Then(`it panic on T#Parallel call`, func(t *testing.T, v *testcase.V) {
					require.True(t, isPanic(func() { t.Parallel() }))
				})
			})

			s.Then(`it will accept T#Parallel call`, func(t *testing.T, v *testcase.V) {
				require.False(t, isPanic(func() { t.Parallel() }))
			})

		})

	})
}

func TestSpec_Let_FallibleValue(t *testing.T) {
	s := testcase.NewSpec(t)

	s.Let(`fallible`, func(v *testcase.V) interface{} {
		return v.T()
	})

	s.Then(`fallible receive the same testing object as this context`, func(t *testing.T, v *testcase.V) {
		t1 := t
		t2 := v.I(`fallible`).(*testing.T)
		require.Equal(t, t1, t2)
	})
}

func TestSpec_LetNow_ValueDefinedAtDeclarationWithoutTheNeedOfFunctionCallback(t *testing.T) {
	t.Skip(`undefined if this worth it, the deep copy reflection approach would be heavy in terms of dependency`)
	//
	//s := testcase.NewSpec(t)
	//
	//s.LetNow(`value`, int(42))
	//
	//s.Then(`LetNow will set a value at the declaration point and not create value during the test execution flow`, func(t *testing.T, v *testcase.V) {
	//	require.Equal(t, int(42), v.I(`value`).(int))
	//})
	//
	//s.When(`pointers used`, func(s *testcase.Spec) {
	//	s.LetNow(`value`, &MyType{Field1: `1`})
	//
	//	s.Then(`tests do not modify the other test case value - A`, func(t *testing.T, v *testcase.V) {
	//		value := v.I(`value`).(*MyType)
	//		require.Equal(t, `1`, value.Field1)
	//		value.Field1 = "A"
	//		require.Equal(t, `A`, v.I(`value`).(*MyType).Field1)
	//	})
	//
	//
	//	s.Then(`tests do not modify the other test case value - B`, func(t *testing.T, v *testcase.V) {
	//		value := v.I(`value`).(*MyType)
	//		require.Equal(t, `1`, value.Field1)
	//		value.Field1 = "B"
	//		require.Equal(t, `B`, v.I(`value`).(*MyType).Field1)
	//	})
	//})
}

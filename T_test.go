package testcase_test

import (
	"math/rand"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/adamluzsi/testcase"
)

func TestT_Let_canBeUsedDuringTest(t *testing.T) {
	s := testcase.NewSpec(t)

	s.Context(`runtime define`, func(s *testcase.Spec) {
		s.Let(`n-original`, func(t *testcase.T) interface{} { return rand.Intn(42) })
		s.Let(`m-original`, func(t *testcase.T) interface{} { return rand.Intn(42) + 100 })

		var exampleMultiReturnFunc = func(t *testcase.T) (int, int) {
			return t.I(`n-original`).(int), t.I(`m-original`).(int)
		}

		s.Context(`Let being set during test runtime`, func(s *testcase.Spec) {
			s.Before(func(t *testcase.T) {
				n, m := exampleMultiReturnFunc(t)
				t.Let(`n`, n)
				t.Let(`m`, m)
			})

			s.Test(`let values which are defined during runtime present in the test`, func(t *testcase.T) {
				require.Equal(t, t.I(`n`), t.I(`n-original`))
				require.Equal(t, t.I(`m`), t.I(`m-original`))
			})
		})
	})

	s.Context(`runtime update`, func(s *testcase.Spec) {
		var initValue = rand.Intn(42)
		s.Let(`x`, func(t *testcase.T) interface{} { return initValue })

		s.Before(func(t *testcase.T) {
			t.Let(`x`, t.I(`x`).(int)+1)
		})

		s.Before(func(t *testcase.T) {
			t.Let(`x`, t.I(`x`).(int)+1)
		})

		s.Test(`let will returns the value then override the runtime variables`, func(t *testcase.T) {
			require.Equal(t, initValue+2, t.I(`x`).(int))
		})
	})

}

func TestT_Defer(t *testing.T) {
	s := testcase.NewSpec(t)

	var res []int

	s.Context(``, func(s *testcase.Spec) {
		s.Before(func(t *testcase.T) {
			res = append(res, 0)
		})

		s.After(func(t *testcase.T) {
			res = append(res, -1)
		})

		s.Context(``, func(s *testcase.Spec) {
			s.Around(func(t *testcase.T) func() {
				res = append(res, 1)
				return func() { res = append(res, -2) }
			})

			s.Context(``, func(s *testcase.Spec) {
				s.Let(`with defer`, func(t *testcase.T) interface{} {
					t.Defer(func() { res = append(res, -3) })
					return 42
				})

				s.Before(func(t *testcase.T) {
					// calling a variable that has defer will ensure
					// that the deferred function call will be executed
					// as part of the *T#defer stack, and not afterwards
					require.Equal(t, 42, t.I(`with defer`).(int))
				})

				s.Test(``, func(t *testcase.T) {
					t.Defer(func() { res = append(res, -4) })
				})
			})
		})
	})

	require.Equal(t, []int{0, 1, -4, -3, -2, -1}, res)
}

func TestT_Defer_withArguments(t *testing.T) {
	s := testcase.NewSpec(t)

	expected := rand.Int() + 1
	var actually int

	type S struct{ ID int }
	s.Let(`value`, func(t *testcase.T) interface{} {
		s := &S{ID: expected}
		t.Defer(func(id int) { actually = id }, s.ID)
		return s
	})

	s.Test(`test that alter the content of value`, func(t *testcase.T) {
		s := t.I(`value`).(*S)
		s.ID = 0
	})

	require.Equal(t, expected, actually)
}

func TestT_Defer_withArgumentsButArgumentCountMismatch(t *testing.T) {
	s := testcase.NewSpec(t)

	s.Let(`value`, func(t *testcase.T) interface{} {
		t.Defer(func(text string) {}, `this would be ok`, `but this extra argument is not ok`)
		return 42
	})

	s.Test(`test that it will panics early on to help ease the pain of seeing mistakes`, func(t *testcase.T) {
		require.Panics(t, func() { _ = t.I(`value`).(int) })
	})

	s.Test(`panic message`, func(t *testcase.T) {
		message := func() (r string) {
			defer func() { r = recover().(string) }()
			_ = t.I(`value`).(int)
			return ``
		}()

		require.Contains(t, message, `/testcase/T_test.go`)
		require.Contains(t, message, `expected 1`)
		require.Contains(t, message, `got 2`)
	})
}

func TestT_Defer_withArgumentsButArgumentTypeMismatch(t *testing.T) {
	s := testcase.NewSpec(t)

	s.Let(`value`, func(t *testcase.T) interface{} {
		t.Defer(func(n int) {}, `this is not ok`)
		return 42
	})

	s.Test(`test that it will panics early on to help ease the pain of seeing mistakes`, func(t *testcase.T) {
		require.Panics(t, func() { _ = t.I(`value`).(int) })
	})

	s.Test(`panic message`, func(t *testcase.T) {
		message := func() (r string) {
			defer func() { r = recover().(string) }()
			_ = t.I(`value`).(int)
			return ``
		}()

		require.Contains(t, message, `/testcase/T_test.go`)
		require.Contains(t, message, `expected int`)
		require.Contains(t, message, `got string`)
	})
}

func TestT_Defer_calledWithoutFunctionAndWillPanic(t *testing.T) {
	testcase.NewSpec(t).Test(`defer expected to panic for non function objects`, func(t *testcase.T) {
		var withReturnValue = func() int { return 42 }

		require.Panics(t, func() { t.Defer(withReturnValue()) })
	})
}

func TestT_Defer_willRunEvenIfSomethingForceTheTestToStopEarly(t *testing.T) {
	s := testcase.NewSpec(t)
	var ran bool
	s.Before(func(t *testcase.T) { t.Defer(func() { ran = true }) })
	s.Test(``, func(t *testcase.T) { t.Skip(`please stop early`) })
	require.True(t, ran)
}

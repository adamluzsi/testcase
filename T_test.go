package testcase_test

import (
	"context"
	"math/rand"
	"sync"
	"testing"
	"time"

	"github.com/adamluzsi/testcase/fixtures"

	"github.com/adamluzsi/testcase/random"

	"github.com/adamluzsi/testcase/internal"

	"github.com/stretchr/testify/require"

	"github.com/adamluzsi/testcase"
)

var _ testing.TB = &testcase.T{}

func TestT_Let_canBeUsedDuringTest(t *testing.T) {
	s := testcase.NewSpec(t)

	s.Context(`runtime define`, func(s *testcase.Spec) {
		s.Let(`n-original`, func(t *testcase.T) interface{} { return rand.Intn(42) })
		s.Let(`m-original`, func(t *testcase.T) interface{} { return rand.Intn(42) + 100 })

		var exampleMultiReturnFunc = func(t *testcase.T) (int, int) {
			return t.I(`n-original`).(int), t.I(`m-original`).(int)
		}

		s.Context(`Let being set during testCase runtime`, func(s *testcase.Spec) {
			s.Before(func(t *testcase.T) {
				n, m := exampleMultiReturnFunc(t)
				t.Set(`n`, n)
				t.Set(`m`, m)
			})

			s.Test(`let values which are defined during runtime present in the testCase`, func(t *testcase.T) {
				require.Equal(t, t.I(`n`), t.I(`n-original`))
				require.Equal(t, t.I(`m`), t.I(`m-original`))
			})
		})
	})

	s.Context(`runtime update`, func(s *testcase.Spec) {
		var initValue = rand.Intn(42)
		s.Let(`x`, func(t *testcase.T) interface{} { return initValue })

		s.Before(func(t *testcase.T) {
			t.Set(`x`, t.I(`x`).(int)+1)
		})

		s.Before(func(t *testcase.T) {
			t.Set(`x`, t.I(`x`).(int)+1)
		})

		s.Test(`let will returns the value then override the runtime vars`, func(t *testcase.T) {
			require.Equal(t, initValue+2, t.I(`x`).(int))
		})
	})

}

func TestT_Defer(t *testing.T) {
	var res []int

	t.Run(``, func(t *testing.T) {
		s := testcase.NewSpec(t)

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
	})

	require.Equal(t, []int{0, 1, -4, -3, -2, -1}, res)
}

// TB#Cleanup https://github.com/golang/go/issues/41355
//goland:noinspection GoDeferGo
func TestT_Defer_failNowWillNotHang(t *testing.T) {
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		defer recover()
		s := testcase.NewSpec(&internal.RecorderTB{})

		s.Before(func(t *testcase.T) {
			t.Defer(func() { t.FailNow() })
		})

		s.Context(``, func(s *testcase.Spec) {
			s.Test(``, func(t *testcase.T) {
				panic(`die`)
			})
		})

		s.Test(``, func(t *testcase.T) {})
	}()
	wg.Wait()
}

func TestT_Defer_whenItIsCalledDuringTestBlock(t *testing.T) {
	var itRan bool
	t.Run(``, func(t *testing.T) {
		s := testcase.NewSpec(t)
		s.Test(``, func(t *testcase.T) { t.Defer(func() { itRan = true }) })
	})
	require.True(t, itRan, `then it is expected to ran`)
}

func TestT_Defer_withArguments(t *testing.T) {
	var (
		expected = rand.Int() + 1
		actually int
	)

	t.Run(``, func(t *testing.T) {
		s := testcase.NewSpec(t)
		type S struct{ ID int }

		s.Let(`value`, func(t *testcase.T) interface{} {
			s := &S{ID: expected}
			t.Defer(func(id int) { actually = id }, s.ID)
			return s
		})

		s.Test(`testCase that alter the content of value`, func(t *testcase.T) {
			s := t.I(`value`).(*S)
			s.ID = 0
		})

		s.Test(`interface type with concrete input must be allowed`, func(t *testcase.T) {
			var fn = func(ctx context.Context) {}
			t.Defer(fn, context.Background())
		})
	})

	require.Equal(t, expected, actually)
}

func TestT_Defer_withArgumentsButArgumentCountMismatch(t *testing.T) {
	s := testcase.NewSpec(t)

	var getPanicMessage = func(fn func()) (r string) {
		defer func() { r, _ = recover().(string) }()
		fn()
		return
	}

	s.Let(`value`, func(t *testcase.T) interface{} {
		t.Defer(func(text string) {}, `this would be ok`, `but this extra argument is not ok`)
		return 42
	})

	s.Test(`testCase that it will panics early on to help ease the pain of seeing mistakes`, func(t *testcase.T) {
		require.Panics(t, func() { _ = t.I(`value`).(int) })
	})

	s.Test(`panic message`, func(t *testcase.T) {
		message := getPanicMessage(func() { _ = t.I(`value`).(int) })
		require.Contains(t, message, `/testcase/T_test.go`)
		require.Contains(t, message, `expected 1`)
		require.Contains(t, message, `got 2`)
	})

	s.Test(`interface type with wrong implementation`, func(t *testcase.T) {
		type notContextForSure struct{}
		var fn = func(ctx context.Context) {}
		require.Panics(t, func() { t.Defer(fn, notContextForSure{}) })
		message := getPanicMessage(func() { t.Defer(fn, notContextForSure{}) })
		require.Contains(t, message, `/testcase/T_test.go`)
		require.Contains(t, message, `doesn't implements context.Context`)
		require.Contains(t, message, `argument[0]`)
	})
}

func TestT_Defer_withArgumentsButArgumentTypeMismatch(t *testing.T) {
	s := testcase.NewSpec(t)

	s.Let(`value`, func(t *testcase.T) interface{} {
		t.Defer(func(n int) {}, `this is not ok`)
		return 42
	})

	s.Test(`testCase that it will panics early on to help ease the pain of seeing mistakes`, func(t *testcase.T) {
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

func TestT_TB(t *testing.T) {
	s := testcase.NewSpec(t)

	for i := 0; i < 10; i++ {
		var ts []testing.TB
		s.Test(`*testcase.TB is set to the given testcase's *testing.T`, func(t *testcase.T) {
			require.NotNil(t, t.TB)
			require.NotContains(t, ts, t.TB, `TB should be unique for each testCase run`)
			ts = append(ts, t.TB)
		})
	}
}

func TestT_Defer_calledWithoutFunctionAndWillPanic(t *testing.T) {
	s := testcase.NewSpec(t)

	s.Test(`defer expected to panic for non function input as first value`, func(t *testcase.T) {
		var withReturnValue = func() int { return 42 }

		require.Panics(t, func() { t.Defer(withReturnValue()) })
	})

	s.Test(`defer expected to panic for invalid inputs`, func(t *testcase.T) {
		var dummyClose = func() error { return nil }

		require.PanicsWithValue(t, `T#Defer can only take functions`, func() { t.Defer(dummyClose()) })
	})

}

func TestT_Defer_willRunEvenIfSomethingForceTheTestToStopEarly(t *testing.T) {
	var ran bool
	t.Run(``, func(t *testing.T) {
		s := testcase.NewSpec(t)
		s.Before(func(t *testcase.T) { t.Defer(func() { ran = true }) })
		s.Test(``, func(t *testcase.T) { t.Skip(`please stop early`) })
	})
	require.True(t, ran)
}

func TestT_HasTag(t *testing.T) {
	s := testcase.NewSpec(t)

	s.Context(`a`, func(s *testcase.Spec) {
		s.Tag(`a`)

		s.Context(`b`, func(s *testcase.Spec) {
			s.Tag(`b`)

			s.Context(`c`, func(s *testcase.Spec) {
				s.Tag(`c`)

				s.Test(`c`, func(t *testcase.T) {
					require.True(t, t.HasTag(`a`))
					require.True(t, t.HasTag(`b`))
					require.True(t, t.HasTag(`c`))
					require.False(t, t.HasTag(`d`))
				})
			})

			s.Test(`b`, func(t *testcase.T) {
				require.True(t, t.HasTag(`a`))
				require.True(t, t.HasTag(`b`))
				require.False(t, t.HasTag(`c`))
				require.False(t, t.HasTag(`d`))
			})
		})

		s.Test(`a`, func(t *testcase.T) {
			require.True(t, t.HasTag(`a`))
			require.False(t, t.HasTag(`b`))
			require.False(t, t.HasTag(`c`))
			require.False(t, t.HasTag(`d`))
		})
	})

	s.Test(``, func(t *testcase.T) {
		require.False(t, t.HasTag(`a`))
		require.False(t, t.HasTag(`b`))
		require.False(t, t.HasTag(`c`))
		require.False(t, t.HasTag(`d`))
	})
}

func TestT_Random(t *testing.T) {
	randomGenerationWorks := func(t *testcase.T) {
		testcase.Retry{Strategy: testcase.Waiter{WaitDuration: time.Second}}.Assert(t, func(tb testing.TB) {
			require.True(tb, 0 < t.Random.Int())
		})
	}

	t.Run(`when environment value is set`, func(t *testing.T) {
		testcase.SetEnv(t, testcase.EnvKeySeed, `42`)
		s := testcase.NewSpec(t)
		s.Test(``, func(t *testcase.T) {
			require.Equal(t, random.New(rand.NewSource(42)), t.Random)

			randomGenerationWorks(t)
		})
	})

	s := testcase.NewSpec(t)
	s.Test(``, func(t *testcase.T) {
		randomGenerationWorks(t)
	})
}

func TestNewT(t *testing.T) {
	y := testcase.Var{Name: "Y"}
	v := testcase.Var{
		Name: "the answer",
		Init: func(t *testcase.T) interface{} {
			return t.Random.Int()
		},
	}
	vGet := func(t *testcase.T) int {
		return v.Get(t).(int)
	}

	tb := &internal.StubTB{}
	t.Cleanup(tb.Finish)
	s := testcase.NewSpec(tb)
	expectedY := fixtures.Random.Int()
	y.LetValue(s, expectedY)
	subject := testcase.NewT(tb, s)
	require.Equal(t, expectedY, y.Get(subject).(int), "use the passed spec's runtime context after set-up")
	require.Equal(tb, vGet(subject), vGet(subject), `has test variable cache`)
}

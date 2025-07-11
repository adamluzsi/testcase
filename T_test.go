package testcase_test

import (
	"context"
	"fmt"
	"math/rand"
	"os"
	"sync/atomic"
	"testing"
	"time"

	"go.llib.dev/testcase/assert"
	"go.llib.dev/testcase/contracts"
	"go.llib.dev/testcase/internal/doubles"
	"go.llib.dev/testcase/internal/environ"
	"go.llib.dev/testcase/sandbox"

	"go.llib.dev/testcase/random"

	"go.llib.dev/testcase"
)

var _ testing.TB = &testcase.T{}

func TestT_implementsTestingTB(t *testing.T) {
	testcase.RunSuite(t, contracts.TestingTB{
		Subject: func(t *testcase.T) testing.TB {
			stub := &doubles.TB{}
			t.Cleanup(stub.Finish)
			return testcase.NewTWithSpec(stub, nil)
		},
	})
}

func TestVar_Set_canBeUsedDuringTest(t *testing.T) {
	s := testcase.NewSpec(t)

	s.Context(`runtime define`, func(s *testcase.Spec) {
		nog := testcase.Let(s, func(t *testcase.T) int { return rand.Intn(42) })
		mog := testcase.Let(s, func(t *testcase.T) int { return rand.Intn(42) + 100 })

		var exampleMultiReturnFunc = func(t *testcase.T) (int, int) {
			return nog.Get(t), mog.Get(t)
		}

		s.Context(`Let being set during testCase runtime`, func(s *testcase.Spec) {
			n := testcase.Var[int]{ID: "n"}
			m := testcase.Var[int]{ID: "m"}

			s.Before(func(t *testcase.T) {
				nv, mv := exampleMultiReturnFunc(t)
				n.Set(t, nv)
				m.Set(t, mv)
			})

			s.Test(`let values which are defined during runtime present in the testCase`, func(t *testcase.T) {
				t.Must.Equal(n.Get(t), nog.Get(t))
				t.Must.Equal(m.Get(t), mog.Get(t))
			})
		})
	})

	s.Context(`runtime update`, func(s *testcase.Spec) {
		var initValue = rand.Intn(42)
		x := testcase.Let(s, func(t *testcase.T) int { return initValue })

		s.Before(func(t *testcase.T) {
			x.Set(t, x.Get(t)+1)
		})

		s.Before(func(t *testcase.T) {
			x.Set(t, x.Get(t)+1)
		})

		s.Test(`let will returns the value then override the runtime vars`, func(t *testcase.T) {
			t.Must.Equal(initValue+2, x.Get(t))
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
					wDefer := testcase.Let(s, func(t *testcase.T) int {
						t.Defer(func() { res = append(res, -3) })
						return 42
					})

					s.Before(func(t *testcase.T) {
						// calling a variable that has defer will ensure
						// that the deferred function call will be executed
						// as part of the *T#defer stack, and not afterwards
						t.Must.Equal(42, wDefer.Get(t))
					})

					s.Test(``, func(t *testcase.T) {
						t.Defer(func() { res = append(res, -4) })
					})
				})
			})
		})
	})

	assert.Must(t).Equal([]int{0, 1, -4, -3, -2, -1}, res)
}

// TB#Cleanup https://github.com/golang/go/issues/41355
//
//goland:noinspection GoDeferGo
func TestT_Defer_failNowWillNotHang(t *testing.T) {
	assert.Within(t, time.Second, func(ctx context.Context) {
		sandbox.Run(func() {
			s := testcase.NewSpec(&doubles.TB{})
			s.Test(``, func(t *testcase.T) {
				t.Defer(func() { t.FailNow() })

				panic(`die`)
			})
		})
	})
}

func TestT_Defer_whenItIsCalledDuringTestBlock(t *testing.T) {
	var itRan bool
	t.Run(``, func(t *testing.T) {
		s := testcase.NewSpec(t)
		s.Test(``, func(t *testcase.T) { t.Defer(func() { itRan = true }) })
	})
	assert.Must(t).True(itRan, `then it is expected to ran`)
}

func TestT_Defer_withArguments(t *testing.T) {
	var (
		expected = rand.Int() + 1
		actually int
	)

	t.Run(``, func(t *testing.T) {
		s := testcase.NewSpec(t)
		type S struct{ ID int }

		v := testcase.Let(s, func(t *testcase.T) *S {
			s := &S{ID: expected}
			t.Defer(func(id int) { actually = id }, s.ID)
			return s
		})

		s.Test(`testCase that alter the content of value`, func(t *testcase.T) {
			v.Get(t).ID = 0
		})

		s.Test(`interface type with concrete input must be allowed`, func(t *testcase.T) {
			var fn = func(ctx context.Context) {}
			t.Defer(fn, context.Background())
		})
	})

	assert.Must(t).Equal(expected, actually)
}

func TestT_Defer_runsOnlyAfterTestIsdone(t *testing.T) {
	s := testcase.NewSpec(t)

	CTX := testcase.Let(s, func(t *testcase.T) func() context.Context {
		return func() context.Context {
			ctx, cancel := context.WithCancel(context.Background())
			t.Cleanup(cancel)
			t.Defer(cancel)
			return ctx
		}
	})

	s.Before(func(t *testcase.T) {
		t.Cleanup(func() { t.Must.NoError(CTX.Get(t)().Err()) })
		t.Defer(func() { t.Must.NoError(CTX.Get(t)().Err()) })
	})

	s.Test("", func(t *testcase.T) {})
}

func TestT_Defer_withArgumentsButArgumentCountMismatch(t *testing.T) {
	s := testcase.NewSpec(t)

	var getPanicMessage = func(fn func()) (r string) {
		defer func() { r, _ = recover().(string) }()
		fn()
		return
	}

	v := testcase.Let(s, func(t *testcase.T) int {
		t.Defer(func(text string) {}, `this would be ok`, `but this extra argument is not ok`)
		return 42
	})

	s.Test(`testCase that it will panics early on to help ease the pain of seeing mistakes`, func(t *testcase.T) {
		t.Must.Panic(func() { _ = v.Get(t) })
	})

	s.Test(`panic message`, func(t *testcase.T) {
		message := getPanicMessage(func() { _ = v.Get(t) })
		t.Must.Contains(message, `/testcase/T_test.go`)
		t.Must.Contains(message, `expected 1`)
		t.Must.Contains(message, `got 2`)
	})

	s.Test(`interface type with wrong implementation`, func(t *testcase.T) {
		type notContextForSure struct{}
		var fn = func(ctx context.Context) {}
		t.Must.Panic(func() { t.Defer(fn, notContextForSure{}) })
		message := getPanicMessage(func() { t.Defer(fn, notContextForSure{}) })
		t.Must.Contains(message, `/testcase/T_test.go`)
		t.Must.Contains(message, `doesn't implements context.Context`)
		t.Must.Contains(message, `argument[0]`)
	})
}

func TestT_Defer_withArgumentsButArgumentTypeMismatch(t *testing.T) {
	s := testcase.NewSpec(t)

	v := testcase.Let(s, func(t *testcase.T) int {
		t.Defer(func(n int) {}, `this is not ok`)
		return 42
	})

	s.Test(`testCase that it will panics early on to help ease the pain of seeing mistakes`, func(t *testcase.T) {
		t.Must.Panic(func() { _ = v.Get(t) })
	})

	s.Test(`panic message`, func(t *testcase.T) {
		message := func() (r string) {
			defer func() { r = recover().(string) }()
			_ = v.Get(t)
			return ``
		}()

		t.Must.Contains(message, `/testcase/T_test.go`)
		t.Must.Contains(message, `expected int`)
		t.Must.Contains(message, `got string`)
	})
}

func TestT_TB(t *testing.T) {
	s := testcase.NewSpec(t)

	for i := 0; i < 10; i++ {
		var ts []testing.TB
		s.Test(`*testcase.TB is set to the given testcase's *testing.T`, func(t *testcase.T) {
			t.Must.NotNil(t.TB)
			t.Must.NotContains(ts, t.TB, `TB should be unique for each testCase run`)
			ts = append(ts, t.TB)
		})
	}
}

func TestT_Defer_calledWithoutFunctionAndWillPanic(t *testing.T) {
	s := testcase.NewSpec(t)

	s.Test(`defer expected to panic for non function input as first value`, func(t *testcase.T) {
		var withReturnValue = func() int { return 42 }

		t.Must.Panic(func() { t.Defer(withReturnValue()) })
	})

	s.Test(`defer expected to panic for invalid inputs`, func(t *testcase.T) {
		var dummyClose = func() error { return nil }
		pv := t.Must.Panic(func() { t.Defer(dummyClose()) })
		t.Must.Contains(pv, `T#Defer can only take functions`)
	})

}

func TestT_Defer_willRunEvenIfSomethingForceTheTestToStopEarly(t *testing.T) {
	var ran bool
	t.Run(``, func(t *testing.T) {
		s := testcase.NewSpec(t)
		s.Before(func(t *testcase.T) { t.Defer(func() { ran = true }) })
		s.Test(``, func(t *testcase.T) { t.Skip(`please stop early`) })
	})
	assert.Must(t).True(ran)
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
					t.Must.True(t.HasTag(`a`))
					t.Must.True(t.HasTag(`b`))
					t.Must.True(t.HasTag(`c`))
					t.Must.True(!t.HasTag(`d`))
				})
			})

			s.Test(`b`, func(t *testcase.T) {
				t.Must.True(t.HasTag(`a`))
				t.Must.True(t.HasTag(`b`))
				t.Must.True(!t.HasTag(`c`))
				t.Must.True(!t.HasTag(`d`))
			})
		})

		s.Test(`a`, func(t *testcase.T) {
			t.Must.True(t.HasTag(`a`))
			t.Must.True(!t.HasTag(`b`))
			t.Must.True(!t.HasTag(`c`))
			t.Must.True(!t.HasTag(`d`))
		})
	})

	s.Test(``, func(t *testcase.T) {
		t.Must.True(!t.HasTag(`a`))
		t.Must.True(!t.HasTag(`b`))
		t.Must.True(!t.HasTag(`c`))
		t.Must.True(!t.HasTag(`d`))
	})
}

func TestT_Random(t *testing.T) {
	randomGenerationWorks := func(t *testcase.T) {
		assert.Retry{Strategy: assert.Waiter{WaitDuration: time.Second}}.Assert(t, func(it testing.TB) {
			assert.True(it, 0 < t.Random.Int())
		})
	}

	t.Run(`when environment value is set`, func(t *testing.T) {
		testcase.SetEnv(t, environ.KeySeed, `42`)
		s := testcase.NewSpec(t)
		s.Test(``, func(t *testcase.T) {
			t.Must.NotEmpty(t.Random)

			randomGenerationWorks(t)
		})
	})

	s := testcase.NewSpec(t)
	s.Test(``, func(t *testcase.T) {
		randomGenerationWorks(t)
	})
}

func TestT_Eventually(t *testing.T) {
	rnd := random.New(random.CryptoSeed{})

	t.Run(`with default eventually retry strategy`, func(t *testing.T) {
		stub := &doubles.TB{}
		s := testcase.NewSpec(stub)
		s.HasSideEffect()
		var eventuallyRan bool
		s.Test(``, func(t *testcase.T) {
			t.Eventually(func(it *testcase.T) {
				eventuallyRan = true
				it.Must.True(t.Random.Bool())
			}) // eventually pass
		})
		stub.Finish()
		s.Finish()
		assert.Must(t).True(!stub.IsFailed, `expected to pass`)
		assert.Must(t).True(eventuallyRan)
	})

	t.Run(`with config passed`, func(t *testing.T) {
		stub := &doubles.TB{}
		var strategyUsed bool
		strategy := assert.LoopFunc(func(condition func() bool) {
			strategyUsed = true
			for condition() {
			}
		})
		s := testcase.NewSpec(stub, testcase.WithRetryStrategy(strategy))
		s.HasSideEffect()
		s.Test(``, func(t *testcase.T) {
			t.Eventually(func(it *testcase.T) {
				it.Must.True(t.Random.Bool())
			}) // eventually pass
		})
		stub.Finish()
		s.Finish()
		assert.Must(t).True(!stub.IsFailed, `expected to pass`)
		assert.Must(t).True(strategyUsed, `retry strategy of the eventually call was used`)
	})

	t.Run("Eventually uses a testcase.T that allows its functionalities all from the the eventually block", func(t *testing.T) {
		// After extensive testing, we discovered that constantly switching between using `assert.It` and `testcase.T` is risky
		// and prone to errors without adding any benefit.
		// Therefore, our next step is to streamline the API to simplify testing.

		stub := &doubles.TB{}
		s := testcase.NewSpec(stub)
		s.HasSideEffect()

		expTag := rnd.StringNC(5, random.CharsetAlpha())
		s.Tag(expTag)

		expVal := rnd.Int()
		v := testcase.LetValue(s, expVal)

		var ran bool
		s.Test(``, func(tcT *testcase.T) {
			tcT.Eventually(func(it *testcase.T) {
				ran = true
				assert.Equal(t, expVal, v.Get(it))
				assert.Equal(t, tcT.Random, it.Random)
				assert.True(t, tcT.HasTag(expTag))
				assert.True(t, tcT.TB != it.TB)
			})
		})

		stub.Finish()
		s.Finish()

		assert.Must(t).True(!stub.IsFailed, `expected to pass`)
		assert.True(t, ran)
	})

	t.Run("smoke", func(t *testing.T) {
		stub := &doubles.TB{}
		s := testcase.NewSpec(stub)
		s.HasSideEffect()

		var ran bool
		s.Test(``, func(tcT *testcase.T) {
			var failed bool
			tcT.Eventually(func(t *testcase.T) {
				if !failed {
					failed = true
					t.FailNow()
				}
				// OK
				ran = true
			})
			assert.False(t, tcT.Failed())
		})

		stub.Finish()
		s.Finish()

		assert.Must(t).True(!stub.IsFailed, `expected to pass`)
		assert.True(t, ran)
	})

	t.Run("when failure occurs during the variable initialisation", func(t *testing.T) {
		t.Run("permanently", func(t *testing.T) {
			stub := &doubles.TB{}
			s := testcase.NewSpec(stub, testcase.WithRetryStrategy(assert.RetryCount(3)))
			s.HasSideEffect()

			v := testcase.Let[int](s, func(t *testcase.T) int {
				t.FailNow() // boom
				return 42
			})

			s.Test(``, func(tcT *testcase.T) {
				tcT.Eventually(func(it *testcase.T) { v.Get(it) })
			})

			stub.Finish()
			s.Finish()

			assert.Must(t).True(stub.IsFailed, `expected to fail`)
		})
		t.Run("temporarily", func(t *testing.T) {
			stub := &doubles.TB{}
			s := testcase.NewSpec(stub)
			s.HasSideEffect()

			failed := testcase.LetValue[bool](s, false)
			counter := testcase.LetValue[int](s, 0)
			expVal := rnd.Int()

			v := testcase.Let[int](s, func(t *testcase.T) int {
				counter.Set(t, counter.Get(t)+1)
				if !failed.Get(t) {
					failed.Set(t, true)
					t.FailNow() // boom
				}
				return expVal
			})

			s.Test(``, func(tcT *testcase.T) {
				tcT.Eventually(func(it *testcase.T) {
					assert.Equal(t, v.Get(it), expVal)
				})
				assert.Equal(t, v.Get(tcT), expVal)
				assert.Equal(t, counter.Get(tcT), 2, "it was expected that the variable init block only run twice, one for failure and one for success")
			})

			stub.Finish()
			s.Finish()

			assert.Must(t).False(stub.IsFailed, `expected to pass`)
		})
	})
}

func ExampleNewTWithSpec() {
	s := testcase.NewSpec(nil)
	// some spec specific configuration
	s.Before(func(t *testcase.T) {})

	var tb testing.TB // placeholder
	tc := testcase.NewTWithSpec(tb, s)
	_ = tc
}

func TestNewTWithSpec(t *testing.T) {
	rnd := random.New(random.CryptoSeed{})
	y := testcase.Var[int]{ID: "Y"}
	v := testcase.Var[int]{
		ID:   "the answer",
		Init: func(t *testcase.T) int { return t.Random.Int() },
	}
	t.Run(`with *Spec`, func(t *testing.T) {
		tb := &doubles.TB{}
		t.Cleanup(tb.Finish)
		s := testcase.NewSpec(tb)
		expectedY := rnd.Int()
		y.LetValue(s, expectedY)
		subject := testcase.NewTWithSpec(tb, s)
		assert.Must(t).Equal(expectedY, y.Get(subject), "use the passed spec's runtime context after set-up")
		assert.Must(t).Equal(v.Get(subject), v.Get(subject), `has test variable cache`)
	})
	t.Run(`without *Spec`, func(t *testing.T) {
		tb := &doubles.TB{}
		t.Cleanup(tb.Finish)
		expectedY := rnd.Int()
		subject := testcase.NewTWithSpec(tb, nil)
		y.Set(subject, expectedY)
		assert.Must(t).Equal(expectedY, y.Get(subject))
		assert.Must(t).Equal(v.Get(subject), v.Get(subject), `has test variable cache`)
	})
	t.Run(`with *testcase.T, same returned`, func(t *testing.T) {
		tb := &doubles.TB{}
		t.Cleanup(tb.Finish)
		tcT1 := testcase.NewTWithSpec(tb, nil)
		tcT2 := testcase.NewTWithSpec(tcT1, nil)
		assert.Must(t).Equal(tcT1, tcT2)
	})
	t.Run(`when nil received, nil is returned`, func(t *testing.T) {
		assert.Must(t).Nil(testcase.NewTWithSpec(nil, nil))
	})
	t.Run(`when NewT is retrieved multiple times, hooks executed only once`, func(t *testing.T) {
		stb := &doubles.TB{}
		s := testcase.NewSpec(stb)
		var out []struct{}
		s.Before(func(t *testcase.T) {
			out = append(out, struct{}{})
		})
		tct := testcase.NewTWithSpec(stb, s)
		tct = testcase.NewTWithSpec(tct, s)
		tct = testcase.NewTWithSpec(tct, s)
		stb.Finish()
		assert.Equal(t, 1, len(out))
	})
}

func ExampleNewT() {
	var tb testing.TB // placeholder
	_ = testcase.NewT(tb)
}

func TestNewT(t *testing.T) {
	rnd := random.New(random.CryptoSeed{})
	y := testcase.Var[int]{ID: "Y"}
	v := testcase.Var[int]{
		ID:   "the answer",
		Init: func(t *testcase.T) int { return t.Random.Int() },
	}
	t.Run(`smoke`, func(t *testing.T) {
		tb := &doubles.TB{}
		t.Cleanup(tb.Finish)
		expectedY := rnd.Int()
		subject := testcase.NewT(tb)
		y.Set(subject, expectedY)
		assert.Must(t).Equal(expectedY, y.Get(subject))
		assert.Must(t).Equal(v.Get(subject), v.Get(subject), `has test variable cache`)
	})
	t.Run(`with *testcase.T, same returned`, func(t *testing.T) {
		tb := &doubles.TB{}
		t.Cleanup(tb.Finish)
		tcT1 := testcase.NewT(tb)
		tcT2 := testcase.NewT(tcT1)
		assert.Must(t).Equal(tcT1, tcT2)
	})
	t.Run(`when nil received, nil is returned`, func(t *testing.T) {
		assert.Must(t).Nil(testcase.NewT(nil))
	})
}

func BenchmarkT_varDoesNotCountTowardsRun(b *testing.B) {
	s := testcase.NewSpec(b)

	ab := testcase.Let(s, func(t *testcase.T) int {
		time.Sleep(time.Second / 2)
		return t.Random.Int()
	})
	bv := testcase.Let(s, func(t *testcase.T) int {
		_ = ab.Get(t)
		time.Sleep(time.Second / 2)
		return t.Random.Int()
	})
	s.Test(`run`, func(t *testcase.T) {
		// if the benchmark subject is too fast
		// the benchmark goes into a really long measuring loop.
		//
		// expected to perform max around ~ 1001000000 ns/op
		time.Sleep(time.Second)
		_ = bv.Get(t)
	})
}

func TestT_SkipUntil(t *testing.T) {
	const timeLayout = "2006-01-02"
	const skipUntilFormat = "Skip time %s"
	const skipExpiredFormat = "[SkipUntil] expired on %s"
	rnd := random.New(rand.NewSource(time.Now().UnixNano()))
	future := time.Now().AddDate(0, 0, 1)
	t.Run("before SkipUntil deadline, test is skipped", func(t *testing.T) {
		stubTB := &doubles.TB{}
		s := testcase.NewSpec(stubTB)
		var ran bool
		s.Test("", func(t *testcase.T) {
			t.SkipUntil(future.Year(), future.Month(), future.Day(), future.Hour())
			ran = true
		})
		sandbox.Run(func() { s.Finish() })
		assert.Must(t).False(ran)
		assert.Must(t).False(stubTB.LastTB().IsFailed)
		assert.Must(t).True(stubTB.LastTB().IsSkipped)
		assert.Must(t).Contains(stubTB.LastTB().Logs.String(), fmt.Sprintf(skipUntilFormat, future.Format(timeLayout)))
	})
	t.Run("at or after SkipUntil deadline, test is failed", func(t *testing.T) {
		stubTB := &doubles.TB{}
		s := testcase.NewSpec(stubTB)
		today := time.Now().AddDate(0, 0, -1*rnd.IntN(3))
		var ran bool
		s.Test("", func(t *testcase.T) {
			t.SkipUntil(today.Year(), today.Month(), today.Day(), today.Hour())
			ran = true
		})
		sandbox.Run(func() { s.Finish() })
		assert.Must(t).True(ran)
		assert.Must(t).False(stubTB.LastTB().IsFailed)
		assert.Must(t).Contains(stubTB.LastTB().Logs.String(), fmt.Sprintf(skipExpiredFormat, today.Format(timeLayout)))
	})
}

func TestT_UnsetEnv(t *testing.T) {
	const key = "TEST_KEY"
	t.Setenv(key, "this")
	s := testcase.NewSpec(t)
	s.HasSideEffect()
	s.Test("on unset", func(t *testcase.T) {
		t.UnsetEnv(key)
		_, ok := os.LookupEnv(key)
		t.Must.False(ok)
	})
	s.Test("when not used", func(t *testcase.T) {
		_, ok := os.LookupEnv(key)
		t.Must.True(ok)
	})
	s.Finish()

	t.Run("on Parallel test", func(t *testing.T) {
		dtb := &doubles.TB{}

		sandbox.Run(func() {
			defer dtb.Finish()
			defer s.Finish()
			s := testcase.NewSpec(dtb)
			s.Parallel()
			s.Test("on unset it will fail", func(t *testcase.T) {
				t.UnsetEnv(key)
			})
		})

		assert.True(t, dtb.IsFailed)
	})
}

func TestT_SetEnv(t *testing.T) {
	const key = "TEST_KEY"
	defaultValue := "this"
	t.Setenv(key, defaultValue)
	s := testcase.NewSpec(t)
	s.HasSideEffect()
	s.Test("on set", func(t *testcase.T) {
		r := t.Random.StringNC(5, random.CharsetAlpha())
		t.SetEnv(key, r)
		val, ok := os.LookupEnv(key)
		t.Must.True(ok)
		t.Must.Equal(r, val)
	})
	s.Test("on not used", func(t *testcase.T) {
		val, ok := os.LookupEnv(key)
		t.Must.True(ok)
		t.Must.Equal(defaultValue, val)
	})
	s.Finish()
}

func TestT_Setenv(t *testing.T) {
	const key = "TEST_KEY"
	defaultValue := "this"
	t.Setenv(key, defaultValue)
	s := testcase.NewSpec(t)
	s.HasSideEffect()
	s.Test("on set", func(t *testcase.T) {
		r := t.Random.StringNC(5, random.CharsetAlpha())
		t.Setenv(key, r)
		val, ok := os.LookupEnv(key)
		t.Must.True(ok)
		t.Must.Equal(r, val)
	})
	s.Test("on not used", func(t *testcase.T) {
		val, ok := os.LookupEnv(key)
		t.Must.True(ok)
		t.Must.Equal(defaultValue, val)
	})
	s.Finish()
}

func TestT_LogPretty(t *testing.T) {
	dtb := &doubles.TB{}
	tct := testcase.ToT(dtb)
	tct.LogPretty([]int{1, 2, 4})
	type X struct{ Foo string }
	tct.LogPretty(X{Foo: "hello"})
	dtb.Finish()
	assert.Contains(t, dtb.Logs.String(), "[]int{\n\t1,\n\t2,\n\t4,\n}")
	assert.Contains(t, dtb.Logs.String(), "testcase_test.X{\n\tFoo: \"hello\",\n}")
}

func ExampleT_Done() {
	s := testcase.NewSpec(nil)

	s.Test("", func(t *testcase.T) {
		go func() {
			select {
			// case do something for the test
			case <-t.Done():
				return // test is over, time to garbage collect
			}
		}()
	})
}

func TestT_Done(t *testing.T) {
	s := testcase.NewSpec(t)

	var isdone = func(t *testcase.T) bool {
		select {
		case <-t.Done():
			return true
		default:
			return false
		}
	}

	var done int32
	s.Test("", func(t *testcase.T) {
		assert.False(t, isdone(t))
		go func() {
			<-t.Done() // after the test is done
			atomic.AddInt32(&done, 1)
		}()
		t.Cleanup(func() {
			assert.False(t, isdone(t),
				"during cleanup the done should be not ready")

			t.Cleanup(func() {
				assert.False(t, isdone(t),
					"during a cleanup of cleanup, done should not be ready")
			})
		})
	})

	s.Finish()

	assert.Eventually(t, time.Second, func(t testing.TB) {
		assert.Equal(t, atomic.LoadInt32(&done), 1)
	})
}

func TestT_OnFail(t *testing.T) {
	t.Run("on success", func(t *testing.T) {
		dtb := &doubles.TB{}
		var done bool
		s := testcase.NewSpec(dtb)
		s.Test("", func(t *testcase.T) {
			t.OnFail(func() { done = true })
		})
		s.Finish()
		dtb.Finish()
		assert.Equal(t, false, done)
	})
	t.Run("on failure", func(t *testing.T) {
		dtb := &doubles.TB{}
		var done bool
		s := testcase.NewSpec(dtb)
		s.Test("", func(t *testcase.T) {
			t.OnFail(func() { done = true })
			t.FailNow()
		})
		s.Finish()
		dtb.Finish()
		assert.Equal(t, true, done)
	})
	t.Run("race", func(t *testing.T) {
		dtb := &doubles.TB{}
		s := testcase.NewSpec(dtb)
		s.Test("", func(t *testcase.T) {
			testcase.Race(func() {
				t.OnFail(func() {})
			}, func() {
				t.OnFail(func() {})
			})
			t.FailNow()
		})
		s.Finish()
		dtb.Finish()
	})
}

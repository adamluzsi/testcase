package testcase_test

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"math/rand"
	"os"
	"sync/atomic"
	"testing"
	"time"

	"go.llib.dev/testcase"
	"go.llib.dev/testcase/assert"
	"go.llib.dev/testcase/internal/doubles"
	"go.llib.dev/testcase/internal/env"
	"go.llib.dev/testcase/internal/example/mydomain"
	"go.llib.dev/testcase/let"
	"go.llib.dev/testcase/random"
	"go.llib.dev/testcase/sandbox"
	"go.llib.dev/testcase/tcsync"
)

func ExampleRace() {
	v := mydomain.MyUseCase{}

	// running `go test` with the `-race` flag should help you detect unsafe implementations.
	// each block run at the same time in a race situation
	testcase.Race(func() {
		v.ThreadSafeCall()
	}, func() {
		v.ThreadSafeCall()
	})
}

func ExampleRace_fromSlice() {
	v := mydomain.MyUseCase{}

	var fns []func()
	for i := 0; i < 7; i++ {
		fns = append(fns, v.ThreadSafeCall)
	}

	testcase.Race(fns...)
}

func TestRace(t *testing.T) {
	s := testcase.NewSpec(t)

	// functions are the lambdas that the subject races against each other.
	functions := testcase.Let[[]func()](s, nil)

	// act runs the functions in a race, wrapped in a sandbox
	// so the propagated goroutine exit can be observed as the outcome.
	act := func(t *testcase.T) sandbox.RunOutcome {
		return sandbox.Run(func() {
			testcase.Race(functions.Get(t)...)
		})
	}

	s.Then(`functions run in race against each other`, func(t *testcase.T) {
		// The shared, non-thread-safe state must be fresh on every retry,
		// so it is rebuilt inside the Eventually block.
		t.Eventually(func(t *testcase.T) {
			var counter, total int32
			expTotal := t.Random.IntBetween(3, 7)
			functions.Set(t, random.Slice(expTotal, func() func() {
				return func() { fnWithRaceCondition(&counter, &total) }
			}))

			act(t)

			assert.Equal(t, int32(expTotal), total)
			t.Log(`counter:`, counter, `total:`, total)
			assert.True(t, counter < total,
				`counter was expected to be less that the total block run during race`)
		})
	})

	s.When(`several functions are provided`, func(s *testcase.Spec) {
		sum := testcase.Let(s, func(t *testcase.T) *int32 {
			return new(int32)
		})
		functions.Let(s, func(t *testcase.T) []func() {
			n := sum.Get(t)
			return []func(){
				func() { atomic.AddInt32(n, 1) },
				func() { atomic.AddInt32(n, 10) },
				func() { atomic.AddInt32(n, 100) },
				func() { atomic.AddInt32(n, 1000) },
			}
		})

		s.Then(`each block runs once`, func(t *testcase.T) {
			act(t)

			assert.Equal(t, int32(1111), atomic.LoadInt32(sum.Get(t)))
		})
	})

	s.When(`one of the lambdas exits its goroutine early, e.g. with FailNow`, func(s *testcase.Spec) {
		var fn1Finished, fn2Finished bool
		functions.Let(s, func(t *testcase.T) []func() {
			fn1Finished, fn2Finished = false, false
			return []func(){
				func() {
					fn1Finished = true
				},
				func() {
					fakeTB := &doubles.TB{}
					// this only meant to represent why goroutine exit needs to be propagated.
					fakeTB.FailNow()
					fn2Finished = true
				},
			}
		})

		s.Then(`goexit is propagated back from the lambdas after each lambda finished`, func(t *testcase.T) {
			out := act(t)

			assert.True(t, fn1Finished, `first race block was expected to finish regardless the second's FailNow call`)
			assert.True(t, !fn2Finished, `second race block exited with FailNow, it shouldn't finished`)
			assert.True(t, out.Goexit, `after the second block exited, the exit should have propagated to the top one`)
		})
	})
}

//go:norace
func fnWithRaceCondition(flakyCounter *int32, correctCounter *int32) {
	atomic.AddInt32(correctCounter, 1)
	c := *flakyCounter // copy
	time.Sleep(time.Millisecond)
	*flakyCounter = c + 1 // counter++ would not work
}

func TestSkipUntil(t *testing.T) {
	const timeLayout = "2006-01-02"
	const skipUntilFormat = "Skip time %s"
	const skipExpiredFormat = "[SkipUntil] expired on %s"
	rnd := random.New(rand.NewSource(time.Now().UnixNano()))
	t.Run("before SkipUntil deadline, test is skipped", func(t *testing.T) {
		stubTB := &doubles.TB{}
		future := time.Now().AddDate(0, 0, 1)
		ro := sandbox.Run(func() {
			testcase.SkipUntil(stubTB, future.Year(), future.Month(), future.Day(), future.Hour())
		})
		assert.False(t, ro.OK)
		assert.True(t, ro.Goexit)
		assert.False(t, stubTB.IsFailed)
		assert.True(t, stubTB.IsSkipped)
		assert.Must(t).Contains(stubTB.Logs.String(), fmt.Sprintf(skipUntilFormat, future.Format(timeLayout)))
	})
	t.Run("SkipUntil won't skip when the deadline reached", func(t *testing.T) {
		stubTB := &doubles.TB{}
		now := time.Now()
		ro := sandbox.Run(func() { testcase.SkipUntil(stubTB, now.Year(), now.Month(), now.Day(), now.Hour()) })
		assert.True(t, ro.OK)
		assert.False(t, ro.Goexit)
		assert.False(t, stubTB.IsFailed)
		assert.False(t, stubTB.IsSkipped)
		assert.Must(t).Contains(stubTB.Logs.String(), fmt.Sprintf(skipExpiredFormat, now.Format(timeLayout)))
	})
	t.Run("at or after SkipUntil deadline, test is failed", func(t *testing.T) {
		stubTB := &doubles.TB{}
		today := time.Now().AddDate(0, 0, -1*rnd.IntN(3))
		ro := sandbox.Run(func() { testcase.SkipUntil(stubTB, today.Year(), today.Month(), today.Day(), today.Hour()) })
		assert.True(t, ro.OK)
		assert.False(t, ro.Goexit)
		assert.False(t, stubTB.IsFailed)
		assert.Must(t).Contains(stubTB.Logs.String(), fmt.Sprintf(skipExpiredFormat, today.Format(timeLayout)))
	})
}

func TestSetEnv(t *testing.T) {
	rnd := random.New(random.CryptoSeed{})
	key := rnd.StringNC(5, random.CharsetAlpha())
	ovalue := rnd.StringNC(5, random.CharsetAlpha())
	env.SetEnv(t, key, ovalue)

	t.Run("on use", func(t *testing.T) {
		var dtb doubles.TB
		defer dtb.Finish()

		nvalue := rnd.StringNC(5, random.CharsetAlpha())
		testcase.SetEnv(&dtb, key, nvalue)

		got, ok := os.LookupEnv(key)
		assert.True(t, ok)
		assert.Equal(t, got, nvalue)

		dtb.Finish()

		got, ok = os.LookupEnv(key)
		assert.True(t, ok)
		assert.Equal(t, got, ovalue)

		assert.Empty(t, dtb.Logs.String())
	})

	t.Run("on not using it", func(t *testing.T) {
		assert.Equal(t, ovalue, os.Getenv(key))
	})

	t.Run("on use when failure occurs", func(t *testing.T) {
		var dtb doubles.TB
		defer dtb.Finish()

		nvalue := rnd.StringNC(5, random.CharsetAlpha())
		testcase.SetEnv(&dtb, key, nvalue)

		dtb.Fail()
		dtb.Finish()

		assert.Contains(t, dtb.Logs.String(), fmt.Sprintf("env %s=%q", key, nvalue))
	})
}

func TestUnsetEnv(t *testing.T) {
	rnd := random.New(random.CryptoSeed{})
	key := rnd.StringNC(5, random.CharsetAlpha())
	value := rnd.StringNC(5, random.CharsetAlpha())
	env.SetEnv(t, key, value)

	t.Run("on use", func(t *testing.T) {
		var dtb doubles.TB
		defer dtb.Finish()

		testcase.UnsetEnv(&dtb, key)

		_, ok := os.LookupEnv(key)
		assert.False(t, ok)

		dtb.Finish()

		_, ok = os.LookupEnv(key)
		assert.True(t, ok)

		assert.Empty(t, dtb.Logs.String())
	})

	t.Run("on not using it", func(t *testing.T) {
		env, ok := os.LookupEnv(key)
		assert.True(t, ok)
		assert.Equal(t, value, env)
	})

	t.Run("on use when failure occurs", func(t *testing.T) {
		var dtb doubles.TB
		defer dtb.Finish()

		testcase.UnsetEnv(&dtb, key)

		dtb.Fail()
		dtb.Finish()

		assert.Contains(t, dtb.Logs.String(), fmt.Sprintf("env unset %s", key))
	})
}

func TestOnFail(t *testing.T) {
	t.Run("happy", func(t *testing.T) {
		var dtb doubles.TB

		var ran bool
		testcase.OnFail(&dtb, func() { ran = true })

		dtb.Finish()

		assert.False(t, ran)
	})
	t.Run("rainy", func(t *testing.T) {
		var dtb doubles.TB

		var ran bool
		testcase.OnFail(&dtb, func() { ran = true })
		dtb.Fail()
		dtb.Finish()

		assert.True(t, ran)
	})
}

func ExampleGetEnv() {
	var tb testing.TB = &testing.T{}
	const EnvKey = "THE_ENV_KEY"

	// get an environment variable, or skip the test
	testcase.GetEnv(tb, EnvKey, tb.SkipNow)

	// get an environment variable, or fail now the test
	testcase.GetEnv(tb, EnvKey, tb.Fail)
	testcase.GetEnv(tb, EnvKey, tb.FailNow)
}

func TestGetEnv(t *testing.T) {
	s := testcase.NewSpec(t)

	var (
		dtb = let.Var(s, func(t *testcase.T) *doubles.TB {
			return &doubles.TB{}
		})
		key = let.Var(s, func(t *testcase.T) string {
			return t.Random.StringNWithCharset(
				t.Random.IntBetween(3, 10),
				random.CharsetAlpha())
		})
	)
	actWithSkip := let.Act(func(t *testcase.T) string {
		return random.Pick(t.Random,
			func() string { return testcase.GetEnv(dtb.Get(t), key.Get(t), dtb.Get(t).SkipNow) },
			func() string { return testcase.GetEnv(dtb.Get(t), key.Get(t)) },
		)()
	})
	actWithFatal := let.Act(func(t *testcase.T) string {
		return random.Pick(t.Random,
			func() string { return testcase.GetEnv(dtb.Get(t), key.Get(t), dtb.Get(t).Fail) },
			func() string { return testcase.GetEnv(dtb.Get(t), key.Get(t), dtb.Get(t).FailNow) },
		)()
	})

	s.When("env variable present in the environment", func(s *testcase.Spec) {
		value := let.String(s)

		s.Before(func(t *testcase.T) {
			testcase.SetEnv(t, key.Get(t), value.Get(t))
		})

		s.Then("value is returned", func(t *testcase.T) {
			assert.Equal(t, value.Get(t), actWithSkip(t))
			assert.Equal(t, value.Get(t), actWithFatal(t))
		})
	})

	s.When("env variable is absent in the environment", func(s *testcase.Spec) {
		s.Before(func(t *testcase.T) {
			testcase.UnsetEnv(t, key.Get(t))
		})

		s.Then("on fail use, fail call is expected", func(t *testcase.T) {
			sandbox.Run(func() { actWithFatal(t) })

			assert.True(t, dtb.Get(t).IsFailed)
		})

		s.Then("on skip use, skip call is expected", func(t *testcase.T) {
			sandbox.Run(func() { actWithSkip(t) })

			assert.True(t, dtb.Get(t).IsSkipped)
		})

		s.Then("on every case, logging of the missing env variable is expected", func(t *testcase.T) {
			sandbox.Run(func() {
				if t.Random.Bool() {
					actWithSkip(t)
				} else {
					actWithFatal(t)
				}
			})

			assert.Contains(t, dtb.Get(t).Logs.String(), key.Get(t))
			assert.Contains(t, dtb.Get(t).Logs.String(), "not found")
		})
	})
}

var Stdout io.Writer = os.Stdout

func FuncWithGlobalDependency(a ...any) {
	fmt.Fprintln(Stdout, a...)
}

func ExampleSetGlobal() {
	var t *testing.T

	t.Run("foo", func(t *testing.T) {
		var buf bytes.Buffer
		testcase.SetGlobal[io.Writer](t, &Stdout, &buf)

		FuncWithGlobalDependency("hello")
		assert.Contains(t, buf.String(), "hello")
	})

	t.Run("bar", func(t *testing.T) {
		var buf bytes.Buffer
		testcase.SetGlobal[io.Writer](t, &Stdout, &buf)

		FuncWithGlobalDependency("world")
		assert.Contains(t, buf.String(), "world")
	})
}

func TestSetGlobal_smoke(t *testing.T) {
	s := testcase.NewSpec(t)

	s.Test("single value single use", func(t *testcase.T) {
		var ftb testcase.FakeTB
		var (
			oldVal = t.Random.Int()
			newVal = t.Random.Int()
		)
		var globVar int = oldVal
		testcase.SetGlobal(&ftb, &globVar, newVal)
		assert.Equal(t, globVar, newVal)
		ftb.Finish()
		assert.Equal(t, globVar, oldVal)
	})

	s.Test("consequential calls on same variable", func(t *testcase.T) {
		var ftb testcase.FakeTB
		var (
			oldVal  = t.Random.Int()
			newVal  = t.Random.Int()
			newVal2 = t.Random.Int()
		)
		var globVar int = oldVal
		testcase.SetGlobal(&ftb, &globVar, newVal)
		assert.Equal(t, globVar, newVal)
		testcase.SetGlobal(&ftb, &globVar, newVal2)
		assert.Equal(t, globVar, newVal2)
		ftb.Finish()
		assert.Equal(t, globVar, oldVal)
	})

	s.Test("the use during parallel execution is not allowed", func(t *testcase.T) {
		var ftb testcase.FakeTB
		var (
			oldVal = t.Random.Int()
			newVal = t.Random.Int()
		)
		var globVar int = oldVal
		testcase.SetGlobal(&ftb, &globVar, newVal)
		assert.Equal(t, globVar, newVal)

		var done tcsync.Phaser
		go func() {
			defer done.Finish()
			var ftb testcase.FakeTB
			defer ftb.Finish()
			var newVal = t.Random.Int()
			o := sandbox.Run(func() {
				testcase.SetGlobal(&ftb, &globVar, newVal)
			})
			assert.Should(t).False(o.OK, "expected that the execution was interrupted")
			assert.Should(t).True(ftb.IsFailed, "expected that the test is marked as failed due to incorrect test arrangment code with SetGlobal")
		}()

		assert.Within(t, time.Second, func(ctx context.Context) {
			done.Wait()
		})

		ftb.Finish()
		assert.Equal(t, globVar, oldVal)
	})
}

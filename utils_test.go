package testcase_test

import (
	"fmt"
	"math/rand"
	"os"
	"testing"
	"time"

	"go.llib.dev/testcase"
	"go.llib.dev/testcase/assert"
	"go.llib.dev/testcase/internal/doubles"
	"go.llib.dev/testcase/internal/env"
	"go.llib.dev/testcase/let"
	"go.llib.dev/testcase/random"
	"go.llib.dev/testcase/sandbox"
)

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
		assert.Must(t).False(ro.OK)
		assert.Must(t).True(ro.Goexit)
		assert.Must(t).False(stubTB.IsFailed)
		assert.Must(t).True(stubTB.IsSkipped)
		assert.Must(t).Contain(stubTB.Logs.String(), fmt.Sprintf(skipUntilFormat, future.Format(timeLayout)))
	})
	t.Run("SkipUntil won't skip when the deadline reached", func(t *testing.T) {
		stubTB := &doubles.TB{}
		now := time.Now()
		ro := sandbox.Run(func() { testcase.SkipUntil(stubTB, now.Year(), now.Month(), now.Day(), now.Hour()) })
		assert.Must(t).True(ro.OK)
		assert.Must(t).False(ro.Goexit)
		assert.Must(t).False(stubTB.IsFailed)
		assert.Must(t).False(stubTB.IsSkipped)
		assert.Must(t).Contain(stubTB.Logs.String(), fmt.Sprintf(skipExpiredFormat, now.Format(timeLayout)))
	})
	t.Run("at or after SkipUntil deadline, test is failed", func(t *testing.T) {
		stubTB := &doubles.TB{}
		today := time.Now().AddDate(0, 0, -1*rnd.IntN(3))
		ro := sandbox.Run(func() { testcase.SkipUntil(stubTB, today.Year(), today.Month(), today.Day(), today.Hour()) })
		assert.Must(t).True(ro.OK)
		assert.Must(t).False(ro.Goexit)
		assert.Must(t).False(stubTB.IsFailed)
		assert.Must(t).Contain(stubTB.Logs.String(), fmt.Sprintf(skipExpiredFormat, today.Format(timeLayout)))
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

		assert.Contain(t, dtb.Logs.String(), fmt.Sprintf("env %s=%q", key, nvalue))
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

		assert.Contain(t, dtb.Logs.String(), fmt.Sprintf("env unset %s", key))
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
	testcase.GetEnv(tb, EnvKey, tb.Skip)
	testcase.GetEnv(tb, EnvKey, tb.SkipNow)

	// get an environment variable, or fail now the test
	testcase.GetEnv(tb, EnvKey, tb.Fatal)
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
		if t.Random.Bool() {
			return testcase.GetEnv(dtb.Get(t), key.Get(t), dtb.Get(t).Skip)
		}
		return testcase.GetEnv(dtb.Get(t), key.Get(t), dtb.Get(t).SkipNow)
	})
	actWithFatal := let.Act(func(t *testcase.T) string {
		switch t.Random.IntBetween(1, 3) {
		case 1:
			return testcase.GetEnv(dtb.Get(t), key.Get(t), dtb.Get(t).Fail)
		case 2:
			return testcase.GetEnv(dtb.Get(t), key.Get(t), dtb.Get(t).FailNow)
		case 3:
			return testcase.GetEnv(dtb.Get(t), key.Get(t), dtb.Get(t).Fatal)
		default:
			panic("implementation error of the test")
		}
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

			assert.Contain(t, dtb.Get(t).Logs.String(), key.Get(t))
			assert.Contain(t, dtb.Get(t).Logs.String(), "not found")
		})
	})
}

package testcase_test

import (
	"fmt"
	"github.com/adamluzsi/testcase"
	"github.com/adamluzsi/testcase/assert"
	"github.com/adamluzsi/testcase/internal/doubles"
	"github.com/adamluzsi/testcase/internal/env"
	"github.com/adamluzsi/testcase/random"
	"github.com/adamluzsi/testcase/sandbox"
	"math/rand"
	"os"
	"testing"
	"time"
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
	value := rnd.StringNC(5, random.CharsetAlpha())
	env.SetEnv(t, key, value)

	t.Run("on use", func(t *testing.T) {
		nvalue := rnd.StringNC(5, random.CharsetAlpha())
		testcase.SetEnv(t, key, nvalue)
		env, ok := os.LookupEnv(key)
		assert.True(t, ok)
		assert.Equal(t, nvalue, env)
	})

	t.Run("on not using it", func(t *testing.T) {
		assert.Equal(t, value, os.Getenv(key))
	})
}

func TestUnsetEnv(t *testing.T) {
	rnd := random.New(random.CryptoSeed{})
	key := rnd.StringNC(5, random.CharsetAlpha())
	value := rnd.StringNC(5, random.CharsetAlpha())
	env.SetEnv(t, key, value)

	t.Run("on use", func(t *testing.T) {
		testcase.UnsetEnv(t, key)
		_, ok := os.LookupEnv(key)
		assert.False(t, ok)
	})

	t.Run("on not using it", func(t *testing.T) {
		env, ok := os.LookupEnv(key)
		assert.True(t, ok)
		assert.Equal(t, value, env)
	})
}

package testcase_test

import (
	"github.com/adamluzsi/testcase/assert"
	"github.com/adamluzsi/testcase/internal/env"
	"github.com/adamluzsi/testcase/random"
	"os"
	"testing"

	"github.com/adamluzsi/testcase"
)

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

package environ_test

import (
	"fmt"
	"testing"

	"go.llib.dev/testcase/assert"
	"go.llib.dev/testcase/internal"
	"go.llib.dev/testcase/internal/env"
	"go.llib.dev/testcase/internal/environ"
	"go.llib.dev/testcase/random"
)

func Test_checkEnvKeys(t *testing.T) {
	t.Run("when invalid testcase env variable is present in the env", func(t *testing.T) {
		out := internal.StubWarn(t)
		rnd := random.New(random.CryptoSeed{})
		key := fmt.Sprintf("TESTCASE_%s", rnd.StringNC(rnd.IntB(0, 10), random.CharsetASCII()))
		env.SetEnv(t, key, rnd.StringNC(5, random.CharsetASCII()))
		environ.CheckEnvKeys()
		assert.NotEmpty(t, out.String())
		assert.Contain(t, out.String(), key)
		assert.Contain(t, out.String(), "typo")
	})
	t.Run("when only valid env variables are present in the env", func(t *testing.T) {
		out := internal.StubWarn(t)
		env.SetEnv(t, environ.KeySeed, "123")
		env.SetEnv(t, environ.KeyOrdering, "defined")
		environ.CheckEnvKeys()
		assert.Empty(t, out.String())
	})
}

package env_test

import (
	"os"
	"testing"

	"github.com/adamluzsi/testcase"
	"github.com/adamluzsi/testcase/internal/doubles"
	"github.com/adamluzsi/testcase/internal/env"
	"github.com/adamluzsi/testcase/random"
	"github.com/adamluzsi/testcase/sandbox"
)

func Test(t *testing.T) {
	s := testcase.NewSpec(t)

	recTB := testcase.Let(s, func(t *testcase.T) *doubles.RecorderTB {
		dtb := &doubles.TB{}
		t.Defer(dtb.Finish)
		return &doubles.RecorderTB{TB: dtb}
	})

	tbCleanupNow := func(t *testcase.T) { recTB.Get(t).CleanupNow() }

	s.Describe(`#SetEnv`, func(s *testcase.Spec) {
		var (
			key = testcase.Let(s, func(t *testcase.T) string {
				return `TESTING_DATA_` + t.Random.StringNWithCharset(5, random.CharsetAlpha())
			})
			value = testcase.Let(s, func(t *testcase.T) string {
				return t.Random.String()
			})
			subject = func(t *testcase.T) {
				env.SetEnv(recTB.Get(t), key.Get(t), value.Get(t))
			}
		)

		s.Before(func(t *testcase.T) {
			t.Cleanup(func() {
				t.Must.Nil(os.Unsetenv(key.Get(t)))
			})
		})

		s.When(`environment key is invalid`, func(s *testcase.Spec) {
			key.LetValue(s, ``)

			s.Then(`it will return with error`, func(t *testcase.T) {
				var finished bool
				sandbox.Run(func() {
					subject(t)
					finished = true
				})
				t.Must.True(!finished)
				t.Must.True(recTB.Get(t).IsFailed)
			})
		})

		s.When(`no environment variable is set before the call`, func(s *testcase.Spec) {
			s.Before(func(t *testcase.T) {
				t.Must.Nil(os.Unsetenv(key.Get(t)))
			})

			s.Then(`value will be set`, func(t *testcase.T) {
				subject(t)

				v, ok := os.LookupEnv(key.Get(t))
				t.Must.True(ok)
				t.Must.Equal(v, value.Get(t))
			})

			s.Then(`value will be unset after Cleanup`, func(t *testcase.T) {
				subject(t)
				tbCleanupNow(t)

				_, ok := os.LookupEnv(key.Get(t))
				t.Must.True(!ok)
			})
		})

		s.When(`environment variable already had a value`, func(s *testcase.Spec) {
			originalValue := testcase.Let(s, func(t *testcase.T) string {
				return t.Random.String()
			})

			s.Before(func(t *testcase.T) {
				t.Must.Nil(os.Setenv(key.Get(t), originalValue.Get(t)))
			})

			s.Then(`new value will be set`, func(t *testcase.T) {
				subject(t)

				v, ok := os.LookupEnv(key.Get(t))
				t.Must.True(ok)
				t.Must.Equal(v, value.Get(t))
			})

			s.Then(`old value will be restored on Cleanup`, func(t *testcase.T) {
				subject(t)
				tbCleanupNow(t)

				v, ok := os.LookupEnv(key.Get(t))
				t.Must.True(ok)
				t.Must.Equal(v, originalValue.Get(t))
			})
		})
	})

	s.Describe(`#UnsetEnv`, func(s *testcase.Spec) {
		var (
			key = testcase.Let(s, func(t *testcase.T) string {
				return `TESTING_DATA_` + t.Random.StringNWithCharset(5, "ABCDEFGHIJKLMNOPQRSTUVWXYZ")
			})
			subject = func(t *testcase.T) {
				env.UnsetEnv(recTB.Get(t), key.Get(t))
			}
		)

		s.Before(func(t *testcase.T) {
			t.Cleanup(func() {
				t.Must.Nil(os.Unsetenv(key.Get(t)))
			})
		})

		s.When(`no environment variable is set before the call`, func(s *testcase.Spec) {
			s.Before(func(t *testcase.T) {
				t.Must.Nil(os.Unsetenv(key.Get(t)))
			})

			s.Then(`value will be unset after Cleanup`, func(t *testcase.T) {
				subject(t)
				tbCleanupNow(t)

				_, ok := os.LookupEnv(key.Get(t))
				t.Must.True(!ok)
			})
		})

		s.When(`environment variable already had a value`, func(s *testcase.Spec) {
			originalValue := testcase.Let(s, func(t *testcase.T) string {
				return t.Random.String()
			})

			s.Before(func(t *testcase.T) {
				t.Must.Nil(os.Setenv(key.Get(t), originalValue.Get(t)))
			})

			s.Then(`os env value will be unset`, func(t *testcase.T) {
				subject(t)

				_, ok := os.LookupEnv(key.Get(t))
				t.Must.True(!ok)
			})

			s.Then(`old value will be restored after the Cleanup`, func(t *testcase.T) {
				subject(t)
				tbCleanupNow(t)

				v, ok := os.LookupEnv(key.Get(t))
				t.Must.True(ok)
				t.Must.Equal(v, originalValue.Get(t))
			})
		})
	})
}

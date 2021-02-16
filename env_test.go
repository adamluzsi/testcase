package testcase_test

import (
	"os"
	"testing"

	"github.com/adamluzsi/testcase"
	"github.com/adamluzsi/testcase/fixtures"
	"github.com/adamluzsi/testcase/internal"
	"github.com/stretchr/testify/require"
)

func TestEnvVarHelpers(t *testing.T) {
	s := testcase.NewSpec(t)
	s.Describe(`#SetEnv`, func(s *testcase.Spec) {
		var (
			tb           = s.Let(`TB`, func(t *testcase.T) interface{} { return &internal.RecorderTB{} })
			tbCleanupNow = func(t *testcase.T) { tb.Get(t).(*internal.RecorderTB).CleanupNow() }
			key          = s.LetValue(`key`, `TESTING_DATA_`+fixtures.Random.String())
			value        = s.LetValue(`value`, fixtures.Random.String())
			subject      = func(t *testcase.T) {
				testcase.SetEnv(tb.Get(t).(testing.TB), key.Get(t).(string), value.Get(t).(string))
			}
		)

		s.After(func(t *testcase.T) {
			require.Nil(t, os.Unsetenv(key.Get(t).(string)))
		})

		s.When(`environment key is invalid`, func(s *testcase.Spec) {
			key.LetValue(s, ``)

			s.Then(`it will return with error`, func(t *testcase.T) {
				var finished bool
				internal.InGoroutine(func() {
					subject(t)
					finished = true
				})
				require.False(t, finished)
				require.True(t, tb.Get(t).(*internal.RecorderTB).IsFailed)
			})
		})

		s.When(`no environment variable is set before the call`, func(s *testcase.Spec) {
			s.Before(func(t *testcase.T) {
				require.Nil(t, os.Unsetenv(key.Get(t).(string)))
			})

			s.Then(`value will be set`, func(t *testcase.T) {
				subject(t)

				v, ok := os.LookupEnv(key.Get(t).(string))
				require.True(t, ok)
				require.Equal(t, v, value.Get(t))
			})

			s.Then(`value will be unset after Cleanup`, func(t *testcase.T) {
				subject(t)
				tbCleanupNow(t)

				_, ok := os.LookupEnv(key.Get(t).(string))
				require.False(t, ok)
			})
		})

		s.When(`environment variable already had a value`, func(s *testcase.Spec) {
			originalValue := s.LetValue(`original value`, fixtures.Random.String())

			s.Before(func(t *testcase.T) {
				require.Nil(t, os.Setenv(key.Get(t).(string), originalValue.Get(t).(string)))
			})

			s.Then(`new value will be set`, func(t *testcase.T) {
				subject(t)

				v, ok := os.LookupEnv(key.Get(t).(string))
				require.True(t, ok)
				require.Equal(t, v, value.Get(t))
			})

			s.Then(`old value will be restored on Cleanup`, func(t *testcase.T) {
				subject(t)
				tbCleanupNow(t)

				v, ok := os.LookupEnv(key.Get(t).(string))
				require.True(t, ok)
				require.Equal(t, v, originalValue.Get(t))
			})
		})
	})

	s.Describe(`#UnsetEnv`, func(s *testcase.Spec) {
		var (
			tb           = s.Let(`TB`, func(t *testcase.T) interface{} { return &internal.RecorderTB{} })
			tbCleanupNow = func(t *testcase.T) { tb.Get(t).(*internal.RecorderTB).CleanupNow() }
			key          = s.LetValue(`key`, `TESTING_DATA_`+fixtures.Random.String())
			subject      = func(t *testcase.T) {
				testcase.UnsetEnv(tb.Get(t).(testing.TB), key.Get(t).(string))
			}
		)

		s.After(func(t *testcase.T) {
			require.Nil(t, os.Unsetenv(key.Get(t).(string)))
		})

		s.When(`no environment variable is set before the call`, func(s *testcase.Spec) {
			s.Before(func(t *testcase.T) {
				require.Nil(t, os.Unsetenv(key.Get(t).(string)))
			})

			s.Then(`value will be unset after Cleanup`, func(t *testcase.T) {
				subject(t)
				tbCleanupNow(t)

				_, ok := os.LookupEnv(key.Get(t).(string))
				require.False(t, ok)
			})
		})

		s.When(`environment variable already had a value`, func(s *testcase.Spec) {
			originalValue := s.LetValue(`original value`, fixtures.Random.String())

			s.Before(func(t *testcase.T) {
				require.Nil(t, os.Setenv(key.Get(t).(string), originalValue.Get(t).(string)))
			})

			s.Then(`os env value will be unset`, func(t *testcase.T) {
				subject(t)

				_, ok := os.LookupEnv(key.Get(t).(string))
				require.False(t, ok)
			})

			s.Then(`old value will be restored after the Cleanup`, func(t *testcase.T) {
				subject(t)
				tbCleanupNow(t)

				v, ok := os.LookupEnv(key.Get(t).(string))
				require.True(t, ok)
				require.Equal(t, v, originalValue.Get(t))
			})
		})
	})
}

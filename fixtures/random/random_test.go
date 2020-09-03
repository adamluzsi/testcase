package random_test

import (
	"fmt"
	"math/rand"
	"testing"
	"time"

	"github.com/adamluzsi/testcase"
	"github.com/adamluzsi/testcase/fixtures/random"

	"github.com/stretchr/testify/require"
)

func TestRandomizer(t *testing.T) {
	s := testcase.NewSpec(t)
	s.Let(`randomizer`, func(t *testcase.T) interface{} {
		return &random.Random{Source: rand.NewSource(time.Now().Unix())}
	})
	s.Let(`source`, func(t *testcase.T) interface{} {
		return rand.NewSource(time.Now().Unix())
	})
	SpecRandomizerMethods(s)
}

func SpecRandomizerMethods(s *testcase.Spec) {
	var randomizer = func(t *testcase.T) *random.Random {
		return t.I(`randomizer`).(*random.Random)
	}

	s.Describe(`Int`, func(s *testcase.Spec) {
		var subject = func(t *testcase.T) int {
			return randomizer(t).Int()
		}

		s.Then(`it returns a non-negative pseudo-random int`, func(t *testcase.T) {
			out := subject(t)
			require.True(t, 0 <= out)
		})

		s.Then(`it returns distinct value on each call`, func(t *testcase.T) {
			require.NotEqual(t, subject(t), subject(t))
		})
	})

	s.Describe(`Float32`, func(s *testcase.Spec) {
		var subject = func(t *testcase.T) float32 {
			return randomizer(t).Float32()
		}

		s.Then(`it returns, as a float32, a pseudo-random number in [0.0,1.0).`, func(t *testcase.T) {
			require.True(t, 0 <= subject(t))
		})

		s.Then(`it returns distinct value on each call`, func(t *testcase.T) {
			require.NotEqual(t, subject(t), subject(t))
		})
	})

	s.Describe(`Float64`, func(s *testcase.Spec) {
		var subject = func(t *testcase.T) float64 {
			return randomizer(t).Float64()
		}

		s.Then(`it returns, as a float64, a pseudo-random number in [0.0,1.0).`, func(t *testcase.T) {
			require.True(t, 0 <= subject(t))
		})

		s.Then(`it returns distinct value on each call`, func(t *testcase.T) {
			require.NotEqual(t, subject(t), subject(t))
		})
	})

	s.Describe(`IntN`, func(s *testcase.Spec) {
		var subject = func(t *testcase.T) int {
			return randomizer(t).IntN(t.I(`n`).(int))
		}

		s.Let(`n`, func(t *testcase.T) interface{} {
			return randomizer(t).IntN(42) + 42 // ensure it is not zero for the test
		})

		s.Test(`returns with random number excluding the received`, func(t *testcase.T) {
			out := subject(t)
			require.True(t, 0 <= out)
			require.True(t, out < t.I(`n`).(int))
		})
	})

	s.Describe(`IntBetween`, func(s *testcase.Spec) {
		var subject = func(t *testcase.T) int {
			return randomizer(t).IntBetween(t.I(`min`).(int), t.I(`max`).(int))
		}

		s.Let(`min`, func(t *testcase.T) interface{} {
			return randomizer(t).IntN(42)
		})

		s.Let(`max`, func(t *testcase.T) interface{} {
			// +1 in the end to ensure that `max` is bigger than `min`
			return randomizer(t).IntN(42) + t.I(`min`).(int) + 1
		})

		s.Then(`it will return a value between the range`, func(t *testcase.T) {
			out := subject(t)
			require.True(t, t.I(`min`).(int) <= out, `expected that from <= than out`)
			require.True(t, out <= t.I(`max`).(int), `expected that out is <= than max`)
		})

		s.And(`min and max is in the negative range`, func(s *testcase.Spec) {
			s.LetValue(`min`, -128)
			s.LetValue(`max`, -64)

			s.Then(`it will return a value between the range`, func(t *testcase.T) {
				out := subject(t)
				require.True(t, t.I(`min`).(int) <= out, `expected that from <= than out`)
				require.True(t, out <= t.I(`max`).(int), `expected that out is <= than max`)
			})
		})

		s.And(`min and max equal`, func(s *testcase.Spec) {
			s.Let(`max`, func(t *testcase.T) interface{} { return t.I(`min`) })

			s.Then(`it returns the min and max value since the range can only have one value`, func(t *testcase.T) {
				require.Equal(t, t.I(`max`), subject(t))
			})
		})
	})

	s.Describe(`ElementFromSlice`, func(s *testcase.Spec) {
		s.Test(`E2E`, func(t *testcase.T) {
			pool := []int{1, 2, 3, 4, 5}
			resSet := make(map[int]struct{})
			for i := 0; i < 1024; i++ {
				res := randomizer(t).ElementFromSlice(pool).(int)
				resSet[res] = struct{}{}
				require.Contains(t, pool, res)
			}
			require.True(t, len(resSet) > 1, fmt.Sprintf(`%#v`, resSet))
		})
	})

	s.Describe(`KeyFromMap`, func(s *testcase.Spec) {
		s.Test(`E2E`, func(t *testcase.T) {
			var keys = []int{1, 2, 3, 4, 5}
			var srcMap = make(map[int]struct{})
			for _, k := range keys {
				srcMap[k] = struct{}{}
			}
			require.Contains(t, keys, randomizer(t).KeyFromMap(srcMap).(int))
		})

		s.Test(`randomness`, func(t *testcase.T) {
			var keys = []int{1, 2, 3, 4, 5}
			var srcMap = make(map[int]struct{})
			for _, k := range keys {
				srcMap[k] = struct{}{}
			}
			resSet := make(map[int]struct{})
			for i := 0; i < 1024; i++ {
				res := randomizer(t).KeyFromMap(srcMap).(int)
				resSet[res] = struct{}{}
				require.Contains(t, keys, res)
			}
			require.True(t, len(resSet) > 1, fmt.Sprintf(`%#v`, resSet))
		})
	})

	s.Describe(`StringN`, func(s *testcase.Spec) {
		var subject = func(t *testcase.T) string {
			return randomizer(t).StringN(t.I(`length`).(int))
		}
		s.Let(`length`, func(t *testcase.T) interface{} {
			return randomizer(t).IntN(42) + 5
		})

		s.Then(`it create a string with a given length`, func(t *testcase.T) {
			require.Equal(t, t.I(`length`).(int), len(subject(t)),
				`it was expected to create string with the given length`)
		})

		s.Then(`it create random strings on each call`, func(t *testcase.T) {
			require.NotEqual(t, subject(t), subject(t),
				`it was expected to create different strings`)
		})
	})

	s.Describe(`StringNWithCharset`, func(s *testcase.Spec) {
		var subject = func(t *testcase.T) string {
			return randomizer(t).StringNWithCharset(t.I(`length`).(int), t.I(`charset`).(string))
		}
		s.Let(`length`, func(t *testcase.T) interface{} {
			return randomizer(t).IntN(42) + 5
		})

		s.Let(`charset`, func(t *testcase.T) interface{} {
			return "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz-"
		})

		s.Then(`it create a string with a given length`, func(t *testcase.T) {
			require.Equal(t, t.I(`length`).(int), len(subject(t)),
				`it was expected to create string with the given length`)
		})

		s.Then(`it create random strings on each call`, func(t *testcase.T) {
			require.NotEqual(t, subject(t), subject(t),
				`it was expected to create different strings`)
		})

		s.Test(`charset defines what characters will be randomly used`, func(t *testcase.T) {
			for _, edge := range []struct {
				charset string
			}{
				{charset: "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz-"},
				{charset: "0123456789"},
				{charset: "ABCDEFGHIJKLMNOPQRSTUVWXYZ"},
				{charset: "-$!/%"},
			} {
				t.Let(`charset`, edge.charset)
				for _, char := range subject(t) {
					require.Contains(t, edge.charset, string(char))
				}
			}
		})
	})

	s.Describe(`Bool`, func(s *testcase.Spec) {
		var subject = func(t *testcase.T) bool {
			return randomizer(t).Bool()
		}

		s.Then(`it return with random bool on each calls`, func(t *testcase.T) {
			var bools = map[bool]struct{}{}
			for i := 0; i <= 1024; i++ {
				bools[subject(t)] = struct{}{}
			}
			require.Equal(t, 2, len(bools))
		})
	})

	s.Describe(`String`, func(s *testcase.Spec) {
		var subject = func(t *testcase.T) string {
			return randomizer(t).String()
		}

		s.Then(`it create strings with different lengths`, func(t *testcase.T) {
			var lengths = make(map[int]struct{})
			for i := 0; i < 1024; i++ {
				lengths[len(subject(t))] = struct{}{}
			}
			require.Greater(t, len(lengths), 1)
		})

		s.Then(`it create random strings on each call`, func(t *testcase.T) {
			require.NotEqual(t, subject(t), subject(t),
				`it was expected to create different strings`)
		})
	})

	s.Describe(`TimeBetween`, func(s *testcase.Spec) {
		var subject = func(t *testcase.T) time.Time {
			return randomizer(t).TimeBetween(t.I(`from`).(time.Time), t.I(`to`).(time.Time))
		}

		s.Let(`from`, func(t *testcase.T) interface{} {
			return time.Now().UTC()
		})

		s.Let(`to`, func(t *testcase.T) interface{} {
			return t.I(`from`).(time.Time).Add(24 * time.Hour)
		})

		s.Then(`it will return a date between the given time range including 'from' and excluding 'to'`, func(t *testcase.T) {
			out := subject(t)
			require.True(t, t.I(`from`).(time.Time).Unix() <= out.Unix(), `expected that from <= than out`)
			require.True(t, out.Unix() < t.I(`to`).(time.Time).Unix(), `expected that out is < than to`)
		})

		s.Then(`it will generate different time on each call`, func(t *testcase.T) {
			require.NotEqual(t, subject(t), subject(t))
		})

		s.And(`from is before 1970-01-01 (unix timestamp 0)`, func(s *testcase.Spec) {
			s.Let(`from`, func(t *testcase.T) interface{} {
				return time.Unix(0, 0).UTC().AddDate(0, -1, 0)
			})

			s.Let(`to`, func(t *testcase.T) interface{} {
				return t.I(`from`).(time.Time).AddDate(0, 0, 1)
			})

			s.Then(`it will generate a random time between 'from' and 'to'`, func(t *testcase.T) {
				out := subject(t)
				require.True(t, t.I(`from`).(time.Time).Unix() <= out.Unix(), `expected that from <= than out`)
				require.True(t, out.Unix() < t.I(`to`).(time.Time).Unix(), `expected that out is < than to`)
			})
		})
	})

	s.Describe(`Time`, func(s *testcase.Spec) {
		var subject = func(t *testcase.T) time.Time {
			return randomizer(t).Time()
		}

		s.Then(`it will generate different time on each call`, func(t *testcase.T) {
			require.NotEqual(t, subject(t), subject(t))
		})
	})
}

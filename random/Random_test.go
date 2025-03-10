package random_test

import (
	"fmt"
	"math/rand"
	"regexp"
	"strings"
	"testing"
	"time"

	"go.llib.dev/testcase/internal"
	"go.llib.dev/testcase/let"

	"go.llib.dev/testcase/random/sextype"

	"go.llib.dev/testcase/assert"
	"go.llib.dev/testcase/random"

	"go.llib.dev/testcase"
)

func TestRandom(t *testing.T) {
	s := testcase.NewSpec(t)
	rnd := testcase.Let(s, func(t *testcase.T) *random.Random {
		return &random.Random{Source: rand.NewSource(time.Now().Unix())}
	})

	SpecRandomMethods(s, rnd)

	s.Context("smoke test", func(s *testcase.Spec) {
		s.Test("randoms are deterministic", func(t *testcase.T) {
			seed := time.Now().Unix()

			rnd.Get(t).Source = rand.NewSource(seed)
			i1 := rnd.Get(t).IntN(42)
			s1 := rnd.Get(t).String()
			t1 := rnd.Get(t).Time()
			u1 := rnd.Get(t).UUID()
			b1 := make([]byte, 42)
			_, _ = rnd.Get(t).Read(b1)

			rnd.Get(t).Source = rand.NewSource(seed)
			i2 := rnd.Get(t).IntN(42)
			s2 := rnd.Get(t).String()
			t2 := rnd.Get(t).Time()
			u2 := rnd.Get(t).UUID()
			b2 := make([]byte, 42)
			_, _ = rnd.Get(t).Read(b2)

			t.Must.Equal(i1, i2)
			t.Must.Equal(s1, s2)
			t.Must.Equal(t1, t2)
			t.Must.Equal(b1, b2)
			t.Must.Equal(u1, u2)
		})
	})
}

func SpecRandomMethods(s *testcase.Spec, rnd testcase.Var[*random.Random]) {
	const SamplingNumber = 1024

	s.Describe(`Int`, func(s *testcase.Spec) {
		var subject = func(t *testcase.T) int {
			return rnd.Get(t).Int()
		}

		s.Then(`it returns a non-negative pseudo-random int`, func(t *testcase.T) {
			out := subject(t)
			assert.Must(t).True(0 <= out)
		})

		s.Then(`it returns distinct value on each call`, func(t *testcase.T) {
			assert.Must(t).NotEqual(subject(t), subject(t))
		})
	})

	s.Describe(`Float32`, func(s *testcase.Spec) {
		var subject = func(t *testcase.T) float32 {
			return rnd.Get(t).Float32()
		}

		s.Then(`it returns, as a float32, a pseudo-random number in [0.0,1.0).`, func(t *testcase.T) {
			assert.Must(t).True(0 <= subject(t))
		})

		s.Then(`it returns distinct value on each call`, func(t *testcase.T) {
			assert.Must(t).NotEqual(subject(t), subject(t))
		})
	})

	s.Describe(`Float64`, func(s *testcase.Spec) {
		var subject = func(t *testcase.T) float64 {
			return rnd.Get(t).Float64()
		}

		s.Then(`it returns, as a float64, a pseudo-random number in [0.0,1.0).`, func(t *testcase.T) {
			assert.Must(t).True(0 <= subject(t))
		})

		s.Then(`it returns distinct value on each call`, func(t *testcase.T) {
			assert.Must(t).NotEqual(subject(t), subject(t))
		})
	})

	s.Describe(`IntN`, func(s *testcase.Spec) {
		n := testcase.Let(s, func(t *testcase.T) int {
			return rnd.Get(t).IntN(42) + 42 // ensure it is not zero for the test
		})
		var subject = func(t *testcase.T) int {
			return rnd.Get(t).IntN(n.Get(t))
		}

		s.Test(`returns with random number excluding the received`, func(t *testcase.T) {
			out := subject(t)
			assert.Must(t).True(0 <= out)
			assert.Must(t).True(out < n.Get(t))
		})
	})

	s.Describe(`IntB`, func(s *testcase.Spec) {
		SpecIntBetween(s, rnd, func(t *testcase.T) func(min, max int) int {
			return rnd.Get(t).IntB
		})
	})

	s.Describe(`IntBetween`, func(s *testcase.Spec) {
		SpecIntBetween(s, rnd, func(t *testcase.T) func(min, max int) int {
			return rnd.Get(t).IntBetween
		})
	})

	s.Describe(`DurationBetween`, func(s *testcase.Spec) {
		SpecDurationBetween(s, rnd, func(t *testcase.T) func(min, max time.Duration) time.Duration {
			return rnd.Get(t).DurationBetween
		})
	})

	s.Describe(`Pick`, func(s *testcase.Spec) {
		s.Test(`E2E`, func(t *testcase.T) {
			pool := []int{1, 2, 3, 4, 5}
			resSet := make(map[int]struct{})
			for i := 0; i < SamplingNumber; i++ {
				res := rnd.Get(t).Pick(pool).(int)
				resSet[res] = struct{}{}
				t.Must.Contain(pool, res)
			}
			assert.Must(t).True(len(resSet) > 1, assert.Message(fmt.Sprintf(`%#v`, resSet)))
		})
	})

	s.Describe(`StringN`, func(s *testcase.Spec) {
		length := testcase.Let(s, func(t *testcase.T) int {
			return rnd.Get(t).IntN(42) + 5
		})
		var subject = func(t *testcase.T) string {
			return rnd.Get(t).StringN(length.Get(t))
		}

		s.Then(`it create a string with a given length`, func(t *testcase.T) {
			t.Must.Equal(length.Get(t), len(subject(t)),
				`it was expected to create string with the given length`)
		})

		s.Then(`it create random strings on each call`, func(t *testcase.T) {
			assert.Must(t).NotEqual(subject(t), subject(t),
				`it was expected to create different strings`)
		})
	})

	s.Describe(`StringNC`, func(s *testcase.Spec) {
		SpecStringNWithCharset(s, rnd, func(t *testcase.T, rnd *random.Random, length int, charset string) string {
			return rnd.StringNC(length, charset)
		})
	})

	s.Describe(`StringNWithCharset`, func(s *testcase.Spec) {
		SpecStringNWithCharset(s, rnd, func(t *testcase.T, rnd *random.Random, length int, charset string) string {
			return rnd.StringNWithCharset(length, charset)
		})
	})

	s.Describe(`Bool`, func(s *testcase.Spec) {
		var subject = func(t *testcase.T) bool {
			return rnd.Get(t).Bool()
		}

		s.Then(`it return with random bool on each calls`, func(t *testcase.T) {
			var bools = map[bool]struct{}{}
			for i := 0; i <= SamplingNumber; i++ {
				bools[subject(t)] = struct{}{}
			}
			t.Must.Equal(2, len(bools))
		})
	})

	s.Describe(`Error`, func(s *testcase.Spec) {
		act := func(t *testcase.T) error {
			return rnd.Get(t).Error()
		}

		s.Then(`it create error with different content`, func(t *testcase.T) {
			var lengths = make(map[string]struct{})
			for i := 0; i < SamplingNumber; i++ {
				err := act(t)
				t.Must.Error(err)
				lengths[err.Error()] = struct{}{}
			}
			t.Must.True(1 < len(lengths))
		})

		s.Then(`it create random errors on each call`, func(t *testcase.T) {
			t.Eventually(func(it *testcase.T) {
				it.Must.NotEqual(act(t), act(t), `it was expected to create different error`)
			})
		})
	})

	s.Describe(`String`, func(s *testcase.Spec) {
		var subject = func(t *testcase.T) string {
			return rnd.Get(t).String()
		}

		s.Then(`it create strings with different lengths`, func(t *testcase.T) {
			var lengths = make(map[int]struct{})
			for i := 0; i < SamplingNumber; i++ {
				lengths[len(subject(t))] = struct{}{}
			}
			t.Must.True(1 < len(lengths))
		})

		s.Then(`it create random strings on each call`, func(t *testcase.T) {
			t.Eventually(func(it *testcase.T) {
				it.Must.NotEqual(subject(t), subject(t), `it was expected to create different strings`)
			})
		})
	})

	s.Describe(`Read`, func(s *testcase.Spec) {
		var (
			p = testcase.Let[[]byte](s, nil)
		)
		act := func(t *testcase.T) (n int, err error) {
			return rnd.Get(t).Read(p.Get(t))
		}

		s.When("input slice is nil", func(s *testcase.Spec) {
			p.Let(s, func(t *testcase.T) []byte {
				return nil
			})

			s.Then("zero read is made", func(t *testcase.T) {
				n, err := act(t)
				t.Must.Nil(err)
				t.Must.Equal(0, n)
			})
		})

		s.When("input slice has a length", func(s *testcase.Spec) {
			length := testcase.Let(s, func(t *testcase.T) int {
				return t.Random.IntB(1, 42)
			})
			p.Let(s, func(t *testcase.T) []byte {
				return make([]byte, length.Get(t))
			})

			s.Then("it reads data equal to the input slice length", func(t *testcase.T) {
				n, err := act(t)
				t.Must.Nil(err)
				t.Must.Equal(length.Get(t), n)
				t.Must.NotEmpty(p.Get(t))
			})

			s.Then("continuous reading yields different results", func(t *testcase.T) {
				sampling := t.Random.IntB(42, 82)
				t.Eventually(func(it *testcase.T) {
					var results = make(map[string]struct{})
					for i := 0; i < sampling; i++ {
						n, err := act(t)
						it.Must.Nil(err)
						it.Must.Equal(length.Get(t), n)
						results[string(p.Get(t))] = struct{}{}
					}
					it.Must.True(1 < len(results), "at least more than one results is expected from a continuous reading")
				})
			}, testcase.Flaky(3))
		})
	})

	s.Describe(`TimeBetween`, func(s *testcase.Spec) {
		SpecTimeBetween(s, rnd, func(t *testcase.T) func(from, to time.Time) time.Time {
			return rnd.Get(t).TimeBetween
		})
	})

	s.Describe(`TimeB`, func(s *testcase.Spec) {
		SpecTimeBetween(s, rnd, func(t *testcase.T) func(from, to time.Time) time.Time {
			return rnd.Get(t).TimeB
		})
	})

	s.Describe(`Time`, func(s *testcase.Spec) {
		var subject = func(t *testcase.T) time.Time {
			return rnd.Get(t).Time()
		}

		s.Then(`it will generate different time on each call`, func(t *testcase.T) {
			assert.Must(t).NotEqual(subject(t), subject(t))
		})

		s.Then(`result is safe to format into RFC3339`, func(t *testcase.T) {
			t1 := subject(t)
			t2, _ := time.Parse(time.RFC3339, t1.Format(time.RFC3339))
			t.Must.Equal(t1.UTC(), t2.UTC())
		})
	})

	s.Describe(`TimeN`, func(s *testcase.Spec) {
		var (
			from = testcase.Let(s, func(t *testcase.T) time.Time {
				return time.Now()
			})
			fromGet = func(t *testcase.T) time.Time { return from.Get(t) }
			years   = testcase.Let(s, func(t *testcase.T) int {
				return t.Random.IntN(42)
			})
			months = testcase.Let(s, func(t *testcase.T) int {
				return t.Random.IntN(42)
			})
			days = testcase.Let(s, func(t *testcase.T) int {
				return t.Random.IntN(42)
			})
		)
		var subject = func(t *testcase.T) time.Time {
			return rnd.Get(t).TimeN(fromGet(t), years.Get(t), months.Get(t), days.Get(t))
		}

		getMaxDate := func(t *testcase.T) time.Time {
			return fromGet(t).AddDate(years.Get(t), months.Get(t), days.Get(t))
		}

		s.Then(`it will return a value greater or equal with "from"`, func(t *testcase.T) {
			t.Must.True(fromGet(t).Unix() <= subject(t).Unix())
		})

		s.Then(`it will return a value less or equal with the maximum expected date that is: "from"+years+months+days`, func(t *testcase.T) {
			t.Must.True(subject(t).Unix() <= getMaxDate(t).Unix())
		})

		s.And(`years is negative`, func(s *testcase.Spec) {
			years.Let(s, func(t *testcase.T) int {
				return t.Random.IntN(42) * -1
			})
			months.Let(s, func(t *testcase.T) int {
				return t.Random.IntN(12) * -1
			})
			days.Let(s, func(t *testcase.T) int {
				return t.Random.IntN(29) * -1
			})

			s.Then(`time shift backwards`, func(t *testcase.T) {
				t.Must.True(subject(t).Unix() <= fromGet(t).Unix())
				t.Must.True(getMaxDate(t).Unix() <= subject(t).Unix())
			})
		})

		s.Then(`stress test`, func(t *testcase.T) {
			min := fromGet(t).Unix()
			max := getMaxDate(t).Unix()
			for i := 0; i < 42; i++ {
				sub := subject(t).Unix()
				t.Must.True(min <= sub)
				t.Must.True(sub <= max)
			}
		})

		s.Then(`result is safe to format into RFC3339`, func(t *testcase.T) {
			t1 := subject(t)
			t2, _ := time.Parse(time.RFC3339, t1.Format(time.RFC3339))
			t.Log("t1:", t1.UnixNano(), "t2:", t2.UnixNano())
			t.Must.Equal(t1.UTC(), t2.UTC())
		})

		s.Then(`using it is race safe`, func(t *testcase.T) {
			rdz := rnd.Get(t)
			f := fromGet(t)
			y := years.Get(t)
			m := months.Get(t)
			d := days.Get(t)
			blk := func() { rdz.TimeN(f, y, m, d) }
			testcase.Race(blk, blk, blk)
		})
	})

	s.Describe("UUID", func(s *testcase.Spec) {
		var act = func(t *testcase.T) string {
			return rnd.Get(t).UUID()
		}

		s.Then("it generates a string that looks like UUID", func(t *testcase.T) {
			const uuidPattern = `^[0-9a-fA-F]{8}\b-[0-9a-fA-F]{4}\b-[0-9a-fA-F]{4}\b-[0-9a-fA-F]{4}\b-[0-9a-fA-F]{12}$`
			rgx, err := regexp.Compile(uuidPattern)
			t.Must.NoError(err)
			t.Must.True(rgx.MatchString(act(t)))
		})

		s.Then("it generates random results on every call", func(t *testcase.T) {
			t.Must.NotEqual(act(t), act(t))
		})

		s.Then("calling it bulk yields relatively random UUIDs", func(t *testcase.T) {
			sampling := t.Random.IntB(512, SamplingNumber)
			t.Eventually(func(it *testcase.T) {
				results := make(map[string]struct{})
				for i := 0; i < sampling; i++ {
					results[act(t)] = struct{}{}
				}
				it.Must.Equal(sampling, len(results))
			})
		})
	})

	s.Describe(".Contact", func(s *testcase.Spec) {
		opts := testcase.LetValue[[]internal.ContactOption](s, nil)
		act := func(t *testcase.T) random.Contact {
			return rnd.Get(t).Contact(opts.Get(t)...)
		}

		s.Context(".FirstName", func(s *testcase.Spec) {

			const (
				exampleFemaleName = "Angela"
				exampleMaleName   = "Adam"
			)

			s.Then("a non empty name is returned", func(t *testcase.T) {
				c := act(t)
				t.Must.NotEmpty(c)
				t.Must.NotEmpty(c.FirstName)
				t.Must.NotEmpty(c.LastName)
				t.Must.NotEmpty(c.Email)
			})

			s.Then("it occasionally returns a valid male name", func(t *testcase.T) {
				t.Eventually(func(it *testcase.T) {
					it.Must.Equal(exampleMaleName, act(t).FirstName)
				})
			})

			s.Then("it occasionally returns a valid female name", func(t *testcase.T) {
				t.Eventually(func(it *testcase.T) {
					it.Must.Equal(exampleFemaleName, act(t).FirstName)
				})
			})

			s.When("male sex type is provided", func(s *testcase.Spec) {
				opts.Let(s, func(t *testcase.T) []internal.ContactOption {
					return []internal.ContactOption{sextype.Male}
				})

				s.Then("it occasionally returns a valid male name", func(t *testcase.T) {
					t.Eventually(func(it *testcase.T) {
						it.Must.Equal(exampleMaleName, act(t).FirstName)
					})
				})

				s.Then("it never returns a female name", func(t *testcase.T) {
					name := rnd.Get(t).Contact(sextype.Female).FirstName
					t.Must.AnyOf(func(a *assert.A) {
						for i := 0; i < SamplingNumber; i++ {
							a.Case(func(it assert.It) {
								it.Must.NotEqual(name, act(t).FirstName)
							})
						}
					})
				})
			})

			s.When("female sex type is provided", func(s *testcase.Spec) {
				opts.Let(s, func(t *testcase.T) []internal.ContactOption {
					return []internal.ContactOption{sextype.Female}
				})

				s.Then("it occasionally returns a valid female name", func(t *testcase.T) {
					t.Eventually(func(it *testcase.T) {
						it.Must.Equal(exampleFemaleName, act(t).FirstName)
					})
				})

				s.Then("it never returns a male name", func(t *testcase.T) {
					name := rnd.Get(t).Contact(sextype.Male).FirstName
					t.Must.AnyOf(func(a *assert.A) {
						for i := 0; i < SamplingNumber; i++ {
							a.Case(func(it assert.It) {
								it.Must.NotEqual(name, act(t).FirstName)
							})
						}
					})
				})
			})

			s.When("both sex type is provided", func(s *testcase.Spec) {
				opts.Let(s, func(t *testcase.T) []internal.ContactOption {
					return []internal.ContactOption{sextype.Female, sextype.Male}
				})

				s.Then("it occasionally returns a valid male name", func(t *testcase.T) {
					t.Eventually(func(it *testcase.T) {
						it.Must.Equal(exampleMaleName, act(t).FirstName)
					})
				})

				s.Then("it occasionally returns a valid female name", func(t *testcase.T) {
					t.Eventually(func(it *testcase.T) {
						it.Must.Equal(exampleFemaleName, act(t).FirstName)
					})
				})
			})
		})

		s.Context(".LastName", func(s *testcase.Spec) {
			s.Then("a non empty name is returned", func(t *testcase.T) {
				t.Must.NotEmpty(act(t).LastName)
			})

			s.Then("it returns a valid common last name", func(t *testcase.T) {
				const exampleLastName = "Walker"

				t.Eventually(func(it *testcase.T) {
					it.Must.Equal(exampleLastName, act(t).LastName)
				})
			})
		})

		s.Context(".Email", func(s *testcase.Spec) {
			s.Then("a non empty name is returned", func(t *testcase.T) {
				t.Must.NotEmpty(act(t).Email)
			})

			s.Then("it returns a valid common email domain", func(t *testcase.T) {
				const exampleDomainSuffix = "@gmail.com"

				t.Eventually(func(it *testcase.T) {
					it.Must.True(strings.HasSuffix(act(t).Email, exampleDomainSuffix))
				})
			})
		})
	})

	s.Describe(".Repeat", func(s *testcase.Spec) {
		var (
			min = let.IntB(s, 5, 7)
			max = let.IntB(s, 12, 42)

			times = testcase.LetValue(s, 0)
			blk   = testcase.Let(s, func(t *testcase.T) func() {
				return func() { times.Set(t, times.Get(t)+1) }
			})
		)
		act := func(t *testcase.T) int {
			return rnd.Get(t).Repeat(min.Get(t), max.Get(t), blk.Get(t))
		}

		s.Then("the number of callback execution will be between the min and the max", func(t *testcase.T) {
			act(t)
			t.Must.True(min.Get(t) <= times.Get(t))
			t.Must.True(times.Get(t) <= max.Get(t))
		})

		s.Then("the number of callback execution will be a random number", func(t *testcase.T) {
			runCounts := make(map[int]struct{})

			for i := 0; i < SamplingNumber; i++ {
				times.Set(t, 0)
				act(t)
				runCounts[times.Get(t)] = struct{}{}
			}

			t.Must.True(1 < len(runCounts))
		})

		s.Then("the number of callback execution is the equal to the one reported back by the act", func(t *testcase.T) {
			got := act(t)
			t.Must.Equal(times.Get(t), got)
		})
	})

	s.Describe(".Domain", func(s *testcase.Spec) {
		act := func(t *testcase.T) string {
			return rnd.Get(t).Domain()
		}

		s.Then("a non empty domain is returned", func(t *testcase.T) {
			t.Must.NotEmpty(act(t))
		})

		s.Then("it returns a valid common domain", func(t *testcase.T) {
			t.Eventually(func(it *testcase.T) { it.Must.Equal(act(t), "google.com") })
			t.Eventually(func(it *testcase.T) { it.Must.Equal(act(t), "amazon.com") })
			t.Eventually(func(it *testcase.T) { it.Must.Equal(act(t), "youtube.com") })
		})
	})

	s.Context("DEPRECATED", func(s *testcase.Spec) {
		s.Describe(".Name().First()", func(s *testcase.Spec) {
			act := func(t *testcase.T) string {
				return rnd.Get(t).Name().First()
			}

			const (
				exampleFemaleName = "Angela"
				exampleMaleName   = "Adam"
			)

			s.Then("a non empty name is returned", func(t *testcase.T) {
				t.Must.NotEmpty(act(t))
			})

			s.Then("it occasionally returns a valid male name", func(t *testcase.T) {
				t.Eventually(func(it *testcase.T) {
					it.Must.Equal(exampleMaleName, act(t))
				})
			})

			s.Then("it occasionally returns a valid female name", func(t *testcase.T) {
				t.Eventually(func(it *testcase.T) {
					it.Must.Equal(exampleFemaleName, act(t))
				})
			})

			s.When("male sex type is provided", func(s *testcase.Spec) {
				act := func(t *testcase.T) string {
					return rnd.Get(t).Name().First(sextype.Male)
				}

				s.Then("it occasionally returns a valid male name", func(t *testcase.T) {
					t.Eventually(func(it *testcase.T) {
						it.Must.Equal(exampleMaleName, act(t))
					})
				})

				s.Then("it never returns a female name", func(t *testcase.T) {
					name := rnd.Get(t).Name().First(sextype.Female)
					t.Must.AnyOf(func(a *assert.A) {
						for i := 0; i < SamplingNumber; i++ {
							a.Case(func(it assert.It) {
								it.Must.NotEqual(name, act(t))
							})
						}
					})
				})
			})

			s.When("female sex type is provided", func(s *testcase.Spec) {
				act := func(t *testcase.T) string {
					return rnd.Get(t).Name().First(sextype.Female)
				}

				s.Then("it occasionally returns a valid female name", func(t *testcase.T) {
					t.Eventually(func(it *testcase.T) {
						it.Must.Equal(exampleFemaleName, act(t))
					})
				})

				s.Then("it never returns a male name", func(t *testcase.T) {
					name := rnd.Get(t).Name().First(sextype.Male)
					t.Must.AnyOf(func(a *assert.A) {
						for i := 0; i < SamplingNumber; i++ {
							a.Case(func(it assert.It) {
								it.Must.NotEqual(name, act(t))
							})
						}
					})
				})
			})

			s.When("both sex type is provided", func(s *testcase.Spec) {
				act := func(t *testcase.T) string {
					return rnd.Get(t).Name().First(sextype.Female, sextype.Male)
				}

				s.Then("it occasionally returns a valid male name", func(t *testcase.T) {
					t.Eventually(func(it *testcase.T) {
						it.Must.Equal(exampleMaleName, act(t))
					})
				})

				s.Then("it occasionally returns a valid female name", func(t *testcase.T) {
					t.Eventually(func(it *testcase.T) {
						it.Must.Equal(exampleFemaleName, act(t))
					})
				})
			})
		})

		s.Describe(".Name().Last()", func(s *testcase.Spec) {
			act := func(t *testcase.T) string {
				return rnd.Get(t).Name().Last()
			}

			s.Then("a non empty name is returned", func(t *testcase.T) {
				t.Must.NotEmpty(act(t))
			})

			s.Then("it returns a valid common last name", func(t *testcase.T) {
				const exampleLastName = "Walker"

				t.Eventually(func(it *testcase.T) {
					it.Must.Equal(exampleLastName, act(t))
				})
			})
		})

		s.Describe(".Email", func(s *testcase.Spec) {
			act := func(t *testcase.T) string {
				return rnd.Get(t).Email()
			}

			s.Then("a non empty name is returned", func(t *testcase.T) {
				t.Must.NotEmpty(act(t))
			})

			s.Then("it returns a valid common email domain", func(t *testcase.T) {
				const exampleDomainSuffix = "@gmail.com"

				t.Eventually(func(it *testcase.T) {
					it.Must.True(strings.HasSuffix(act(t), exampleDomainSuffix))
				})
			})
		})
	})
}

func SpecStringNWithCharset(s *testcase.Spec, rnd testcase.Var[*random.Random], act func(t *testcase.T, rnd *random.Random, length int, charset string) string) {
	length := testcase.Let(s, func(t *testcase.T) int {
		return rnd.Get(t).IntN(42) + 5
	})
	charset := testcase.Let(s, func(t *testcase.T) string {
		return "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz-"
	})
	subject := func(t *testcase.T) string {
		return act(t, rnd.Get(t), length.Get(t), charset.Get(t))
	}

	s.Then(`it create a string with a given length`, func(t *testcase.T) {
		t.Must.Equal(length.Get(t), len(subject(t)),
			`it was expected to create string with the given length`)
	})

	s.Then(`it create random strings on each call`, func(t *testcase.T) {
		assert.Must(t).NotEqual(subject(t), subject(t),
			`it was expected to create different strings`)
	})

	s.Test(`charsetAlpha defines what characters will be randomly used`, func(t *testcase.T) {
		for _, edge := range []struct {
			charset string
		}{
			{charset: "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz-"},
			{charset: "0123456789"},
			{charset: "ABCDEFGHIJKLMNOPQRSTUVWXYZ"},
			{charset: "-$!/%"},
		} {
			charset.Set(t, edge.charset)
			for _, char := range subject(t) {
				t.Must.Contain(edge.charset, string(char))
			}
		}
	})
}

func SpecIntBetween(s *testcase.Spec, rnd testcase.Var[*random.Random], sbj func(*testcase.T) func(min, max int) int) {
	var (
		min = testcase.Let(s, func(t *testcase.T) int {
			return rnd.Get(t).IntN(42)
		})
		max = testcase.Let(s, func(t *testcase.T) int {
			// +1 in the end to ensure that `max` is bigger than `min`
			return rnd.Get(t).IntN(42) + min.Get(t) + 1
		})
		subject = func(t *testcase.T) int {
			return sbj(t)(min.Get(t), max.Get(t))
		}
	)

	s.Then(`it will return a value between the range`, func(t *testcase.T) {
		out := subject(t)
		assert.Must(t).True(min.Get(t) <= out, `expected that from <= than out`)
		assert.Must(t).True(out <= max.Get(t), `expected that out is <= than max`)
	})

	s.And(`min and max is in the negative range`, func(s *testcase.Spec) {
		min.LetValue(s, -128)
		max.LetValue(s, -64)

		s.Then(`it will return a value between the range`, func(t *testcase.T) {
			out := subject(t)
			assert.Must(t).True(min.Get(t) <= out, `expected that from <= than out`)
			assert.Must(t).True(out <= max.Get(t), `expected that out is <= than max`)
		})
	})

	s.And(`min and max equal`, func(s *testcase.Spec) {
		max.Let(s, func(t *testcase.T) int { return min.Get(t) })

		s.Then(`it returns the min and max value since the range can only have one value`, func(t *testcase.T) {
			t.Must.Equal(max.Get(t), subject(t))
		})
	})
}

func SpecDurationBetween(s *testcase.Spec, rnd testcase.Var[*random.Random], sbj func(*testcase.T) func(min, max time.Duration) time.Duration) {
	var (
		min = testcase.Let(s, func(t *testcase.T) time.Duration {
			return time.Duration(rnd.Get(t).IntN(42))
		})
		max = testcase.Let(s, func(t *testcase.T) time.Duration {
			// +1 in the end to ensure that `max` is bigger than `min`
			return time.Duration(rnd.Get(t).IntN(42)) + min.Get(t) + 1
		})
		subject = func(t *testcase.T) time.Duration {
			return sbj(t)(min.Get(t), max.Get(t))
		}
	)

	s.Then(`it will return a value between the range`, func(t *testcase.T) {
		out := subject(t)
		assert.Must(t).True(min.Get(t) <= out, `expected that from <= than out`)
		assert.Must(t).True(out <= max.Get(t), `expected that out is <= than max`)
	})

	s.And(`min and max is in the negative range`, func(s *testcase.Spec) {
		min.LetValue(s, -128)
		max.LetValue(s, -64)

		s.Then(`it will return a value between the range`, func(t *testcase.T) {
			out := subject(t)
			assert.Must(t).True(min.Get(t) <= out, `expected that from <= than out`)
			assert.Must(t).True(out <= max.Get(t), `expected that out is <= than max`)
		})
	})

	s.And(`min and max equal`, func(s *testcase.Spec) {
		max.Let(s, func(t *testcase.T) time.Duration { return min.Get(t) })

		s.Then(`it returns the min and max value since the range can only have one value`, func(t *testcase.T) {
			t.Must.Equal(max.Get(t), subject(t))
		})
	})
}

func SpecTimeBetween(s *testcase.Spec, rnd testcase.Var[*random.Random], sbj func(*testcase.T) func(from, to time.Time) time.Time) {
	fromTime := testcase.Let(s, func(t *testcase.T) time.Time {
		return time.Now().UTC()
	})
	toTime := testcase.Let(s, func(t *testcase.T) time.Time {
		return fromTime.Get(t).Add(24 * time.Hour)
	})
	var subject = func(t *testcase.T) time.Time {
		return sbj(t)(fromTime.Get(t), toTime.Get(t))
	}

	s.Then(`it will return a date between the given time range including 'from' and excluding 'to'`, func(t *testcase.T) {
		out := subject(t)
		assert.Must(t).True(fromTime.Get(t).Unix() <= out.Unix(), `expected that from <= than out`)
		assert.Must(t).True(out.Unix() < toTime.Get(t).Unix(), `expected that out is < than to`)
	})

	s.Then(`it will generate different time on each call`, func(t *testcase.T) {
		assert.Must(t).NotEqual(subject(t), subject(t))
	})

	s.And(`from is before 1970-01-01 (unix timestamp 0)`, func(s *testcase.Spec) {
		fromTime.Let(s, func(t *testcase.T) time.Time {
			return time.Unix(0, 0).UTC().AddDate(0, -1, 0)
		})
		toTime.Let(s, func(t *testcase.T) time.Time {
			return fromTime.Get(t).AddDate(0, 0, 1)
		})

		s.Then(`it will generate a random time between 'from' and 'to'`, func(t *testcase.T) {
			out := subject(t)
			assert.Must(t).True(fromTime.Get(t).Unix() <= out.Unix(), `expected that from <= than out`)
			assert.Must(t).True(out.Unix() < toTime.Get(t).Unix(), `expected that out is < than to`)
		})
	})

	s.Then(`result is safe to format into RFC3339`, func(t *testcase.T) {
		t1 := subject(t)
		t2, _ := time.Parse(time.RFC3339, t1.Format(time.RFC3339))
		t.Must.Equal(t1.UTC(), t2.UTC())
	})

	s.And("till is smaller than from", func(s *testcase.Spec) {
		fromTime.Let(s, func(t *testcase.T) time.Time {
			return time.Date(2000, 1, 1, 12, 0, 0, 0, time.Local)
		})
		toTime.Let(s, func(t *testcase.T) time.Time {
			return fromTime.Get(t).Add(-1 * time.Second)
		})

		s.Then("", func(t *testcase.T) {
			out := assert.Panic(t, func() { subject(t) })
			assert.NotNil(t, out)
			panicMessage := fmt.Sprintf("%v", out)
			assert.Contain(t, panicMessage, `invalid`)
			assert.Contain(t, panicMessage, `[to]`)
			assert.Contain(t, panicMessage, `earlier`)
			assert.Contain(t, panicMessage, `[from]`)
			assert.Contain(t, panicMessage, fromTime.Get(t).Format(time.RFC3339))
			assert.Contain(t, panicMessage, toTime.Get(t).Format(time.RFC3339))
		})
	})
}

func ExamplePick_randomValuePicking() {
	// Pick randomly from the values of 1,2,3
	var _ = random.Pick(nil, 1, 2, 3)
}

func ExamplePick_pseudoRandomValuePicking() {
	// Pick pseudo randomly from the given values using the seed.
	// This will make picking deterministically random when the same seed is used.
	const seed = 42
	rnd := random.New(rand.NewSource(seed))
	var _ = random.Pick(rnd, "one", "two", "three")
}

func TestPick(t *testing.T) {
	s := testcase.NewSpec(t)

	var (
		rnd = testcase.Let[*random.Random](s, nil)
		vs  = testcase.Let(s, func(t *testcase.T) []int {
			return random.Slice(t.Random.IntB(3, 5), t.Random.Int)
		})
	)
	act := func(t *testcase.T) int {
		return random.Pick[int](rnd.Get(t), vs.Get(t)...)
	}

	thenItWillStillSelectARandomValue := func(s *testcase.Spec) {
		s.Then("it will still select a random value", func(t *testcase.T) {
			var exp = make(map[int]struct{})
			for _, k := range vs.Get(t) {
				exp[k] = struct{}{}
			}

			var got = make(map[int]struct{})
			t.Eventually(func(it *testcase.T) {
				got[act(t)] = struct{}{}

				it.Must.ContainExactly(exp, got)
			})
		})
	}

	s.When("random.Random is nil", func(s *testcase.Spec) {
		rnd.LetValue(s, nil)

		thenItWillStillSelectARandomValue(s)
	})

	s.When("random.Random is supplied", func(s *testcase.Spec) {
		seed := let.IntB(s, 0, 42)
		mkSource := func(t *testcase.T) rand.Source {
			return rand.NewSource(int64(seed.Get(t)))
		}
		rnd.Let(s, func(t *testcase.T) *random.Random {
			return random.New(mkSource(t))
		})

		thenItWillStillSelectARandomValue(s)

		s.Then("random pick is determinstic through controlling the seed", func(t *testcase.T) {
			exp := act(t)
			rnd.Get(t).Source = mkSource(t)
			got := act(t)
			t.Must.Equal(exp, got)
		})
	})
}

package random_test

import (
	"reflect"
	"testing"
	"time"

	"github.com/adamluzsi/testcase"
	"github.com/adamluzsi/testcase/assert"
	"github.com/adamluzsi/testcase/random"
)

func TestFactory(t *testing.T) {
	s := testcase.NewSpec(t)

	rnd := testcase.Let(s, func(t *testcase.T) *random.Random {
		return random.New(random.CryptoSeed{})
	})

	factory := testcase.Let(s, func(t *testcase.T) *random.Factory {
		return &random.Factory{}
	})

	s.Describe(`.Make`, func(s *testcase.Spec) {
		T := testcase.Var[any]{ID: `<T>`}
		subject := func(t *testcase.T) interface{} {
			return factory.Get(t).Make(rnd.Get(t), T.Get(t))
		}

		retry := testcase.Eventually{RetryStrategy: testcase.Waiter{
			WaitTimeout: 5 * time.Second,
		}}

		thenItGeneratesVariousValues := func(s *testcase.Spec) {
			s.Then(`it generates various results`, func(t *testcase.T) {
				retry.Assert(t, func(it assert.It) {
					var values []interface{}
					for i := 0; i < 12; i++ {
						v := subject(t)
						it.Must.NotContain(values, v)
						values = append(values, v)
					}
				})
			})
		}

		andTheTypeIsPointer := func(s *testcase.Spec) {
			s.And(`type T is actually a pointer *T`, func(s *testcase.Spec) {
				s.Before(func(t *testcase.T) {
					// T -> *T
					rtype := reflect.TypeOf(T.Get(t))
					rPtrType := reflect.PtrTo(rtype)
					T.Set(t, reflect.New(rPtrType).Elem().Interface())
					t.Logf(`%T`, T.Get(t))
				})

				s.Then(`it generates various results`, func(t *testcase.T) {
					retry.Assert(t, func(it assert.It) {
						var values []interface{}
						for i := 0; i < 12; i++ {
							ptr := subject(t)
							v := reflect.ValueOf(ptr).Elem().Interface()
							it.Must.NotContain(values, v)
							values = append(values, v)
						}
					})
				})
			})
		}

		hasValue := func(t *testcase.T, blk func(v interface{}) bool) {
			retry.Assert(t, func(it assert.It) {
				it.Must.True(blk(subject(t)))
			})
		}

		s.When(`type is int`, func(s *testcase.Spec) {
			T.LetValue(s, int(0))

			s.Then(`value type is correct`, func(t *testcase.T) {
				_ = subject(t).(int)
			})

			s.Then(`non zero value generated`, func(t *testcase.T) {
				hasValue(t, func(v interface{}) bool {
					return v.(int) != 0
				})
			})

			thenItGeneratesVariousValues(s)
			andTheTypeIsPointer(s)
		})

		s.When(`type is int8`, func(s *testcase.Spec) {
			T.LetValue(s, int8(0))

			s.Then(`value type is correct`, func(t *testcase.T) {
				_ = subject(t).(int8)
			})

			s.Then(`non zero value generated`, func(t *testcase.T) {
				hasValue(t, func(v interface{}) bool {
					return v.(int8) != 0
				})
			})

			thenItGeneratesVariousValues(s)
			andTheTypeIsPointer(s)
		})

		s.When(`type is int16`, func(s *testcase.Spec) {
			T.LetValue(s, int16(0))

			s.Then(`value type is correct`, func(t *testcase.T) {
				_ = subject(t).(int16)
			})

			s.Then(`non zero value generated`, func(t *testcase.T) {
				hasValue(t, func(v interface{}) bool {
					return v.(int16) != 0
				})
			})

			thenItGeneratesVariousValues(s)
			andTheTypeIsPointer(s)
		})

		s.When(`type is int32`, func(s *testcase.Spec) {
			T.LetValue(s, int32(0))

			s.Then(`value type is correct`, func(t *testcase.T) {
				_ = subject(t).(int32)
			})

			s.Then(`non zero value generated`, func(t *testcase.T) {
				hasValue(t, func(v interface{}) bool {
					return v.(int32) != 0
				})
			})

			thenItGeneratesVariousValues(s)
			andTheTypeIsPointer(s)
		})

		s.When(`type is int64`, func(s *testcase.Spec) {
			T.LetValue(s, int64(0))

			s.Then(`value type is correct`, func(t *testcase.T) {
				_ = subject(t).(int64)
			})

			s.Then(`non zero value generated`, func(t *testcase.T) {
				hasValue(t, func(v interface{}) bool {
					return v.(int64) != 0
				})
			})

			thenItGeneratesVariousValues(s)
			andTheTypeIsPointer(s)
		})

		s.When(`type is uint`, func(s *testcase.Spec) {
			T.LetValue(s, uint(0))

			s.Then(`value type is correct`, func(t *testcase.T) {
				_ = subject(t).(uint)
			})

			s.Then(`non zero value generated`, func(t *testcase.T) {
				hasValue(t, func(v interface{}) bool {
					return v.(uint) != 0
				})
			})

			thenItGeneratesVariousValues(s)
			andTheTypeIsPointer(s)
		})

		s.When(`type is uint8`, func(s *testcase.Spec) {
			T.LetValue(s, uint8(0))

			s.Then(`value type is correct`, func(t *testcase.T) {
				_ = subject(t).(uint8)
			})

			s.Then(`non zero value generated`, func(t *testcase.T) {
				hasValue(t, func(v interface{}) bool {
					return v.(uint8) != 0
				})
			})

			thenItGeneratesVariousValues(s)
			andTheTypeIsPointer(s)
		})

		s.When(`type is uint16`, func(s *testcase.Spec) {
			T.LetValue(s, uint16(0))

			s.Then(`value type is correct`, func(t *testcase.T) {
				_ = subject(t).(uint16)
			})

			s.Then(`non zero value generated`, func(t *testcase.T) {
				hasValue(t, func(v interface{}) bool {
					return v.(uint16) != 0
				})
			})

			thenItGeneratesVariousValues(s)
			andTheTypeIsPointer(s)
		})

		s.When(`type is uint32`, func(s *testcase.Spec) {
			T.LetValue(s, uint32(0))

			s.Then(`value type is correct`, func(t *testcase.T) {
				_ = subject(t).(uint32)
			})

			s.Then(`non zero value generated`, func(t *testcase.T) {
				hasValue(t, func(v interface{}) bool {
					return v.(uint32) != 0
				})
			})

			thenItGeneratesVariousValues(s)
			andTheTypeIsPointer(s)
		})

		s.When(`type is uint64`, func(s *testcase.Spec) {
			T.LetValue(s, uint64(0))

			s.Then(`value type is correct`, func(t *testcase.T) {
				_ = subject(t).(uint64)
			})

			s.Then(`non zero value generated`, func(t *testcase.T) {
				hasValue(t, func(v interface{}) bool {
					return v.(uint64) != 0
				})
			})

			thenItGeneratesVariousValues(s)
			andTheTypeIsPointer(s)
		})

		s.When(`type is float32`, func(s *testcase.Spec) {
			T.LetValue(s, float32(0))

			s.Then(`value type is correct`, func(t *testcase.T) {
				_ = subject(t).(float32)
			})

			s.Then(`non zero value generated`, func(t *testcase.T) {
				hasValue(t, func(v interface{}) bool {
					return v.(float32) != 0
				})
			})

			thenItGeneratesVariousValues(s)
			andTheTypeIsPointer(s)
		})

		s.When(`type is float64`, func(s *testcase.Spec) {
			T.LetValue(s, float64(0))

			s.Then(`value type is correct`, func(t *testcase.T) {
				_ = subject(t).(float64)
			})

			s.Then(`non zero value generated`, func(t *testcase.T) {
				hasValue(t, func(v interface{}) bool {
					return v.(float64) != 0
				})
			})

			thenItGeneratesVariousValues(s)
			andTheTypeIsPointer(s)
		})

		s.When(`type is time.Time`, func(s *testcase.Spec) {
			T.Let(s, func(t *testcase.T) any {
				return time.Time{}
			})

			s.Then(`value type is correct`, func(t *testcase.T) {
				_ = subject(t).(time.Time)
			})

			s.Then(`non zero value generated`, func(t *testcase.T) {
				hasValue(t, func(v interface{}) bool {
					return !v.(time.Time).IsZero()
				})
			})

			thenItGeneratesVariousValues(s)
			andTheTypeIsPointer(s)
		})

		s.When(`type is time.Duration`, func(s *testcase.Spec) {
			T.LetValue(s, time.Duration(0))

			s.Then(`value type is correct`, func(t *testcase.T) {
				_ = subject(t).(time.Duration)
			})

			s.Then(`non zero value generated`, func(t *testcase.T) {
				hasValue(t, func(v interface{}) bool {
					return v.(time.Duration) != 0
				})
			})

			thenItGeneratesVariousValues(s)
			andTheTypeIsPointer(s)
		})

		s.When(`type is bool`, func(s *testcase.Spec) {
			T.LetValue(s, false)

			s.Then(`value type is correct`, func(t *testcase.T) {
				_ = subject(t).(bool) // assert it's bool
			})

			s.Then(`not just false (zero) value is returned`, func(t *testcase.T) {
				hasValue(t, func(v interface{}) bool {
					return !v.(bool)
				})
			})

			s.Then(`it generates both true and false randomly`, func(t *testcase.T) {
				res := make(map[bool]struct{})
				for i := 0; i < 128; i++ {
					v := subject(t).(bool)
					res[v] = struct{}{}
				}

				_, hasTrue := res[true]
				assert.Must(t).True(hasTrue, `should have true in the generated outputs`)

				_, hasFalse := res[true]
				assert.Must(t).True(hasFalse, `should have false in the generated outputs`)
			})
		})

		s.When(`type is string`, func(s *testcase.Spec) {
			T.LetValue(s, "")

			s.Then(`value type is correct`, func(t *testcase.T) {
				v := subject(t).(string)
				t.Must.True(0 < len(v))
			})

			thenItGeneratesVariousValues(s)
			andTheTypeIsPointer(s)
		})

		s.When(`type is struct`, func(s *testcase.Spec) {
			type Y struct {
				Foo int
				Bar string
				Baz bool
			}
			T.Let(s, func(t *testcase.T) any {
				return Y{}
			})

			s.Then(`value type is correct`, func(t *testcase.T) {
				_ = subject(t).(Y)
			})

			s.Then(`each field receive generated value`, func(t *testcase.T) {
				var hasFoo, hasBar, hasBaz bool

				for i := 0; i < 128; i++ {
					y := subject(t).(Y)

					if y.Foo != 0 {
						hasFoo = true
					}

					if y.Bar != "" {
						hasBar = true
					}
					if y.Baz {
						hasBaz = true
					}
				}

				assert.Must(t).True(hasFoo, `excepted to generate value for Foo`)
				assert.Must(t).True(hasBar, `excepted to generate value for Bar`)
				assert.Must(t).True(hasBaz, `excepted to generate value for Baz`)
			})

			thenItGeneratesVariousValues(s)
			andTheTypeIsPointer(s)
		})

		s.When(`type is map`, func(s *testcase.Spec) {
			T.Let(s, func(t *testcase.T) any {
				return map[string]int{}
			})

			s.Then(`value type is correct`, func(t *testcase.T) {
				_ = subject(t).(map[string]int)
			})

			s.Then(`it will create populated map`, func(t *testcase.T) {
				hasValue(t, func(v interface{}) bool {
					m := v.(map[string]int)
					if len(m) == 0 {
						return false
					}

					for k, v := range m {
						if len(k) == 0 {
							return false
						}

						if v == 0 {
							return false
						}
					}

					return true
				})
			})

			thenItGeneratesVariousValues(s)
			andTheTypeIsPointer(s)
		})

		s.When(`type is slice`, func(s *testcase.Spec) {
			T.Let(s, func(t *testcase.T) any {
				return []string{}
			})

			s.Then(`value type is correct`, func(t *testcase.T) {
				_ = subject(t).([]string)
			})

			s.Then(`it will create populated map`, func(t *testcase.T) {
				hasValue(t, func(v interface{}) bool {
					return 0 < len(v.([]string))
				})
			})

			thenItGeneratesVariousValues(s)
			andTheTypeIsPointer(s)
		})

		s.When(`type is array`, func(s *testcase.Spec) {
			T.Let(s, func(t *testcase.T) any {
				return [13]string{}
			})

			s.Then(`value type is correct`, func(t *testcase.T) {
				_ = subject(t).([13]string)
			})

			s.Then(`it will create populated map`, func(t *testcase.T) {
				hasValue(t, func(v interface{}) bool {
					for _, e := range v.([13]string) {
						if len(e) != 0 {
							return true
						}
					}
					return false
				})
			})

			thenItGeneratesVariousValues(s)
			andTheTypeIsPointer(s)
		})

		s.When(`type is chan`, func(s *testcase.Spec) {
			T.Let(s, func(t *testcase.T) any {
				return make(chan int)
			})

			s.Then(`a not nil channel is created`, func(t *testcase.T) {
				assert.Must(t).NotNil(subject(t).(chan int))
			})
		})

		s.When(`type is nil`, func(s *testcase.Spec) {
			T.Let(s, func(t *testcase.T) any {
				return nil
			})

			s.Then(`it will fail the test`, func(t *testcase.T) {
				t.Log(`nil doesn't make sense as a type value`)
				assert.Must(t).Panic(func() { subject(t) })
			})
		})
	})

	s.Test(`.RegisterType`, func(t *testcase.T) {
		type CustomType struct {
			Foo int
			Bar int
		}

		ff := factory.Get(t)

		ff.RegisterType(CustomType{}, func(rnd *random.Random) any {
			return CustomType{Foo: 42, Bar: rnd.Int()}
		})

		ct := ff.Make(rnd.Get(t), CustomType{}).(CustomType)
		t.Must.Equal(42, ct.Foo)
		t.Must.NotEmpty(ct.Bar)
	})
}

func TestFactoryMake_race(t *testing.T) {
	var (
		rnd = random.New(random.CryptoSeed{})
		ff  = &random.Factory{}
	)
	testcase.Race(
		func() { _ = ff.Make(rnd, int(0)).(int) },
		func() { _ = ff.Make(rnd, int(0)).(int) },
		func() { _ = ff.Make(rnd, int(0)).(int) },
		func() { _ = ff.Make(rnd, int(0)).(int) },
	)
}

// BenchmarkFactory
//
// Conclusion
//
// While optimizing the Factory's initialization could double the speed for a single Make call
// It doesn't worth it as Factory is initialized per test case execution rather than Make call.
func BenchmarkFactory(b *testing.B) {
	rnd := random.New(random.CryptoSeed{})

	const makeCount = 2
	b.Run("cached", func(b *testing.B) {
		f := &random.Factory{}
		b.ResetTimer()

		for i := 0; i < b.N; i++ {
			for i := 0; i < makeCount; i++ {
				_ = f.Make(rnd, int(0))
			}
		}
	})
	b.Run("clean", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			f := &random.Factory{}
			for i := 0; i < makeCount; i++ {
				_ = f.Make(rnd, int(0))
			}
		}
	})
}

package fixtures_test

import (
	"context"
	"reflect"
	"runtime"
	"sync"
	"testing"
	"time"

	"github.com/adamluzsi/testcase"
	"github.com/adamluzsi/testcase/fixtures"
	"github.com/stretchr/testify/require"
)

type FixtureFactory interface {
	Fixture(T interface{}, ctx context.Context) (_T interface{})
	RegisterType(T interface{}, constructor func(context.Context) (T interface{}))
}

var _ FixtureFactory = (*fixtures.Factory)(nil)

func TestFactory(t *testing.T) {
	s := testcase.NewSpec(t)

	factory := s.Let(`*fixture.Factory`, func(t *testcase.T) interface{} {
		return &fixtures.Factory{}
	})
	factoryGet := func(t *testcase.T) *fixtures.Factory {
		return factory.Get(t).(*fixtures.Factory)
	}

	s.Describe(`.Fixture`, func(s *testcase.Spec) {
		T := testcase.Var{Name: `<T>`}
		ctx := s.Let(`ctx`, func(t *testcase.T) interface{} {
			return context.Background()
		})
		subject := func(t *testcase.T) interface{} {
			return factoryGet(t).Fixture(T.Get(t), ctx.Get(t).(context.Context))
		}

		retry := testcase.Retry{Strategy: testcase.Waiter{
			WaitTimeout: 5 * time.Second,
		}}

		thenItGeneratesVariousValues := func(s *testcase.Spec) {
			s.Then(`it generates various results`, func(t *testcase.T) {
				retry.Assert(t, func(tb testing.TB) {
					var values []interface{}
					for i := 0; i < 12; i++ {
						v := subject(t)
						require.NotContains(tb, values, v)
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
					retry.Assert(t, func(tb testing.TB) {
						var values []interface{}
						for i := 0; i < 12; i++ {
							ptr := subject(t)
							v := reflect.ValueOf(ptr).Elem().Interface()
							require.NotContains(tb, values, v)
							values = append(values, v)
						}
					})
				})
			})
		}

		hasValue := func(t *testcase.T, assert func(v interface{}) bool) {
			retry.Assert(t, func(tb testing.TB) {
				require.True(tb, assert(subject(t)))
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
			T.Let(s, func(t *testcase.T) interface{} {
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
				require.True(t, hasTrue, `should have true in the generated outputs`)

				_, hasFalse := res[true]
				require.True(t, hasFalse, `should have false in the generated outputs`)
			})
		})

		s.When(`type is string`, func(s *testcase.Spec) {
			T.LetValue(s, "")

			s.Then(`value type is correct`, func(t *testcase.T) {
				v := subject(t).(string)
				require.NotZero(t, v)
				require.NotEmpty(t, v)
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
			T.Let(s, func(t *testcase.T) interface{} {
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

				require.True(t, hasFoo, `excepted to generate value for Foo`)
				require.True(t, hasBar, `excepted to generate value for Bar`)
				require.True(t, hasBaz, `excepted to generate value for Baz`)
			})

			thenItGeneratesVariousValues(s)
			andTheTypeIsPointer(s)
		})

		s.When(`type is map`, func(s *testcase.Spec) {
			T.Let(s, func(t *testcase.T) interface{} {
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
			T.Let(s, func(t *testcase.T) interface{} {
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
			T.Let(s, func(t *testcase.T) interface{} {
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
			T.Let(s, func(t *testcase.T) interface{} {
				return make(chan int)
			})

			s.Then(`a not nil channel is created`, func(t *testcase.T) {
				require.NotNil(t, subject(t).(chan int))
			})
		})

		s.When(`type is nil`, func(s *testcase.Spec) {
			T.Let(s, func(t *testcase.T) interface{} {
				return nil
			})

			s.Then(`it will fail the test`, func(t *testcase.T) {
				require.Panics(t, func() { subject(t) })
			})
		})
	})

	s.Test(`.RegisterType`, func(t *testcase.T) {
		type CustomType struct {
			Foo int
		}

		ff := factoryGet(t)
		expectedCtx := context.WithValue(context.Background(), "foo", "bar")
		ff.RegisterType(CustomType{}, func(actualCtx context.Context) interface{} {
			require.Equal(t, expectedCtx, actualCtx)
			return CustomType{Foo: 42}
		})

		ct := ff.Fixture(CustomType{}, expectedCtx).(CustomType)
		require.Equal(t, 42, ct.Foo)
	})
}

func TestFactory_spike(t *testing.T) {
	rtime := reflect.TypeOf(time.Now())
	rint64 := reflect.TypeOf(int64(42))
	t.Log(`rtime == rint64`, rtime == rint64)
}

func TestFactory_whenNilRandomInitIsThreadSafe(t *testing.T) {
	var (
		ff    = &fixtures.Factory{}
		start sync.WaitGroup
		wg    sync.WaitGroup
	)
	start.Add(1)

	for i := 0; i < runtime.NumCPU(); i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			start.Wait() // get ready
			_ = ff.Fixture(42, context.Background()).(int)
		}()
	}

	start.Done() // race!
	wg.Wait()    // wait till the end of the race
}

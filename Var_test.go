package testcase_test

import (
	"testing"
	"time"

	"github.com/adamluzsi/testcase"
	"github.com/adamluzsi/testcase/fixtures"
	"github.com/stretchr/testify/require"
)

func TestVar(t *testing.T) {
	s := testcase.NewSpec(t)

	// to testCase testVar, I need side effect by resetting the expected in a before hook
	// the var of testVar needs to be leaked into the testing subjects,
	// and I can't use a testVar to testCase testVar because I need this expected at spec level as well.
	// So to testCase testcase.Var, I can't use fully testcase.Var.
	// This should not be the case for anything else outside of the testing framework.
	s.HasSideEffect()
	var testVar = testcase.Var{Name: fixtures.Random.String()}
	testVarGet := func(t *testcase.T) int { return testVar.Get(t).(int) }
	expected := fixtures.Random.Int()

	s.Describe(`#Get`, func(s *testcase.Spec) {
		subject := func(t *testcase.T) int {
			return testVarGet(t)
		}

		s.When(`no expected defined in the spec and no init logic provided`, func(s *testcase.Spec) {
			s.Then(`it will panic, and warn about the unknown expected`, func(t *testcase.T) {
				require.Panics(t, func() { subject(t) })
			})
		})

		s.When(`spec has value by testCase runtime Var#Set`, func(s *testcase.Spec) {
			s.Before(func(t *testcase.T) {
				testVar.Set(t, expected)
			})

			s.Then(`the expected is returned`, func(t *testcase.T) {
				require.Equal(t, expected, testVar.Get(t))
			})
		})

		s.When(`spec has value set by Var#Let`, func(s *testcase.Spec) {
			testVar.Let(s, func(t *testcase.T) interface{} {
				return expected
			})

			s.Then(`the expected is returned`, func(t *testcase.T) {
				require.Equal(t, expected, testVar.Get(t))
			})
		})

		s.When(`spec has value set by Var#LetValue`, func(s *testcase.Spec) {
			testVar.LetValue(s, expected)

			s.Then(`the expected is returned`, func(t *testcase.T) {
				require.Equal(t, expected, testVar.Get(t))
			})
		})

		s.When(`spec has value set by Spec#Let using the Var.Name`, func(s *testcase.Spec) {
			s.Let(testVar.Name, func(t *testcase.T) interface{} {
				return expected
			})

			s.Then(`the expected is returned`, func(t *testcase.T) {
				require.Equal(t, expected, testVar.Get(t))
			})
		})

		s.When(`spec has value set by Spec#LetValue using the Var.Name`, func(s *testcase.Spec) {
			s.LetValue(testVar.Name, expected)

			s.Then(`the expected is returned`, func(t *testcase.T) {
				require.Equal(t, expected, testVar.Get(t))
			})
		})

		s.When(`Var#Init is defined`, func(s *testcase.Spec) {
			s.HasSideEffect()
			// WARN: do not use any other hook that manipulates the testVar here
			// else the side effect is not guaranteed
			s.Around(func(t *testcase.T) func() {
				testVar.Init = func(t *testcase.T) interface{} { return expected }
				// reset side effect
				return func() { testVar.Init = nil }
			})

			thenValueIsCached := func(s *testcase.Spec) {
				s.Then(`value is cached`, func(t *testcase.T) {
					values := make(map[int]struct{})
					for i := 0; i < 128; i++ {
						values[testVar.Get(t).(int)] = struct{}{}
					}
					const failReason = `it was expected that the value from the var is deterministic/cached within the testCase lifetime`
					require.True(t, len(values) == 1, failReason)
				})
			}

			s.And(`spec don't have value set in any way`, func(s *testcase.Spec) {
				s.Then(`it will return the Value from init`, func(t *testcase.T) {
					require.Equal(t, expected, testVar.Get(t))
				})

				s.And(`.Init creates a non deterministic value`, func(s *testcase.Spec) {
					s.HasSideEffect()
					testVar.Init = func(t *testcase.T) interface{} { return fixtures.Random.Int() }
					defer func() { testVar.Init = nil }()

					thenValueIsCached(s)
				})
			})

			s.And(`spec have value set in some form`, func(s *testcase.Spec) {
				testVar.LetValue(s, 42)

				s.Then(`the expected is returned`, func(t *testcase.T) {
					require.Equal(t, 42, testVar.Get(t))
				})
			})
		})
	})

	s.Describe(`#Set`, func(s *testcase.Spec) {
		subject := func(t *testcase.T) {
			testVar.Set(t, expected)
		}

		s.When(`subject used`, func(s *testcase.Spec) {
			s.Before(subject)

			s.Then(`it will set value in the current testCase`, func(t *testcase.T) {
				require.Equal(t, expected, testVar.Get(t))
			})
		})

		s.When(`subject is not used`, func(s *testcase.Spec) {
			s.Then(`value will be absent`, func(t *testcase.T) {
				require.Panics(t, func() { testVar.Get(t) })
			})
		})
	})

	s.Describe(`#Let`, func(s *testcase.Spec) {
		subject := func(s *testcase.Spec) {
			testVar.Let(s, func(t *testcase.T) interface{} { return expected })
		}

		s.When(`subject used`, func(s *testcase.Spec) {
			subject(s)

			s.Then(`it will set value in the spec level`, func(t *testcase.T) {
				require.Equal(t, expected, testVar.Get(t))
			})
		})

		s.When(`subject is not used on a clean Spec`, func(s *testcase.Spec) {
			s.Then(`value will be absent`, func(t *testcase.T) {
				require.Panics(t, func() { testVar.Get(t) })
			})
		})
	})

	s.Describe(`#LetValue`, func(s *testcase.Spec) {
		subject := func(s *testcase.Spec) {
			testVar.LetValue(s, expected)
		}

		s.When(`subject used`, func(s *testcase.Spec) {
			subject(s)

			s.Then(`it will set value in the spec level`, func(t *testcase.T) {
				require.Equal(t, expected, testVar.Get(t))
			})
		})

		s.When(`subject is not used on a clean Spec`, func(s *testcase.Spec) {
			s.Then(`value will be absent`, func(t *testcase.T) {
				require.Panics(t, func() { testVar.Get(t) })
			})
		})
	})

	s.Describe(`#EagerLoading`, func(s *testcase.Spec) {
		subject := func(s *testcase.Spec) {
			testVar.EagerLoading(s)
		}

		testVar.Let(s, func(t *testcase.T) interface{} {
			return int(time.Now().UnixNano())
		})

		s.When(`subject used`, func(s *testcase.Spec) {
			subject(s)

			s.Then(`value will be eager loaded`, func(t *testcase.T) {
				now := int(time.Now().UnixNano())
				require.Less(t, testVar.Get(t), now)
			})
		})

		s.When(`subject not used`, func(s *testcase.Spec) {
			s.Then(`value will be lazy loaded`, func(t *testcase.T) {
				now := int(time.Now().UnixNano())
				require.Less(t, now, testVar.Get(t))
			})
		})
	})

	s.Describe(`#OnLet`, func(s *testcase.Spec) {
		s.When(`it is provided`, func(s *testcase.Spec) {
			v := testcase.Var /* int */ {
				Name: `foo`,
				OnLet: func(s *testcase.Spec) {
					s.Tag(`on-let`) // test trough side effect
				},
			}

			s.And(`variable is not bound to Spec`, func(s *testcase.Spec) {
				s.Test(`it will panic on Var.Get`, func(t *testcase.T) {
					require.Panics(t, func() { v.Get(t) })
				})

				s.Test(`it will panic on Var.Set`, func(t *testcase.T) {
					require.Panics(t, func() { v.Get(t) })
				})
			})

			s.And(`variable is bound to Spec with Var.Let`, func(s *testcase.Spec) {
				v.Let(s, func(t *testcase.T) interface{} { return 42 })

				s.Test(`Var.Get returns value`, func(t *testcase.T) {
					require.Equal(t, 42, v.Get(t))
				})

				s.Test(`it will apply the setup in the context`, func(t *testcase.T) {
					require.True(t, t.HasTag(`on-let`))
				})
			})

			s.And(`variable is bound to Spec with Var.LetValue`, func(s *testcase.Spec) {
				v.LetValue(s, 42)

				s.Test(`Var.Get returns value`, func(t *testcase.T) {
					require.Equal(t, 42, v.Get(t))
				})

				s.Test(`it will apply the setup in the context`, func(t *testcase.T) {
					require.True(t, t.HasTag(`on-let`))
				})
			})
		})

		s.When(`it is absent`, func(s *testcase.Spec) {
			v := testcase.Var /* int */ {
				Name: `foo`,
			}

			s.And(`variable is not bound to Spec`, func(s *testcase.Spec) {
				v := testcase.Var /* int */ {
					Name: `foo`,
					Init: func(t *testcase.T) interface{} {
						// required to be used without binding Var to Spec
						return 42
					},
				}

				s.Test(`it will return initialized value on Var.Get`, func(t *testcase.T) {
					require.Equal(t, 42, v.Get(t))
				})
			})

			s.And(`variable is bound to Spec with Var.Let`, func(s *testcase.Spec) {
				v.Let(s, func(t *testcase.T) interface{} { return 42 })

				s.Test(`Var.Get returns value`, func(t *testcase.T) {
					require.Equal(t, 42, v.Get(t))
				})

				s.Test(`no hook, no setup`, func(t *testcase.T) {
					require.False(t, t.HasTag(`on-let`))
				})
			})

			s.And(`variable is bound to Spec with Var.LetValue`, func(s *testcase.Spec) {
				v.LetValue(s, 42)

				s.Test(`Var.Get returns value`, func(t *testcase.T) {
					require.Equal(t, 42, v.Get(t))
				})

				s.Test(`no hook, no setup`, func(t *testcase.T) {
					require.False(t, t.HasTag(`on-let`))
				})
			})
		})
	})
}

func TestVar_smokeTest(t *testing.T) {
	s := testcase.NewSpec(t)
	s.NoSideEffect()

	type Entity struct {
		TS int64
	}

	entity1 := s.Let(`entity 1`, func(t *testcase.T) interface{} {
		return Entity{TS: time.Now().UnixNano()}
	})

	entity2 := s.Let(`entity 2`, func(t *testcase.T) interface{} {
		return Entity{TS: time.Now().UnixNano()}
	})

	s.When(`var is allowed to use lazy loading`, func(s *testcase.Spec) {
		// nothing to do here, lazy loading is the default behavior

		s.Then(`it should be initialized when it is first accessed`, func(t *testcase.T) {
			e1ts := entity1.Get(t).(Entity).TS
			time.Sleep(42 * time.Nanosecond)
			e2ts := entity2.Get(t).(Entity).TS
			require.True(t, e1ts < e2ts)
		})
	})

	s.When(`var eager loading is requested`, func(s *testcase.Spec) {
		entity2.EagerLoading(s)

		s.Then(`the value should be evaluated `, func(t *testcase.T) {
			e1ts := entity1.Get(t).(Entity).TS
			time.Sleep(42 * time.Nanosecond)
			t.Log(`now we access entity 2,`)
			t.Log(`but the value should already be evaluated by the time the testCase case block is reached`)
			e2ts := entity2.Get(t).(Entity).TS
			require.True(t, e2ts < e1ts)
		})
	})

	s.When(`var override done at spec spec level`, func(s *testcase.Spec) {
		entity1.Let(s, func(t *testcase.T) interface{} {
			return Entity{TS: 0}
		})

		s.Then(`in the testCase case the overridden value will be the initial value`, func(t *testcase.T) {
			require.True(t, entity1.Get(t).(Entity).TS == 0)
		})

		s.Context(``, func(s *testcase.Spec) {
			entity1.Let(s, func(t *testcase.T) interface{} {
				// defined at spec level -> will be initial value with lazy load
				return Entity{TS: time.Now().UnixNano()}
			})
			s.Before(func(t *testcase.T) {
				// defined at testCase run time, will be eager loaded
				entity2.Set(t, Entity{TS: time.Now().UnixNano()})
			})

			s.Test(`spec level definition should be the `, func(t *testcase.T) {

			})
		})
	})

	s.When(`var override done at testCase runtime level`, func(s *testcase.Spec) {
		s.Before(func(t *testcase.T) {
			entity1.Set(t, Entity{TS: 0})
		})

		s.Then(``, func(t *testcase.T) {
			require.True(t, entity1.Get(t).(Entity).TS == 0)
		})
	})

	s.Context(`var override at testCase runtime level`, func(s *testcase.Spec) {
		s.Before(func(t *testcase.T) {
			entity1.Set(t, Entity{TS: 0})
		})

		s.Then(``, func(t *testcase.T) {
			require.True(t, entity1.Get(t).(Entity).TS == 0)
		})
	})

	s.When(`init block defined for the variable`, func(s *testcase.Spec) {
		entity3 := testcase.Var{
			Name: "entity 4",
			Init: func(t *testcase.T) interface{} {
				return Entity{TS: 42}
			},
		}

		s.And(`var is bound to a spec without providing a let variable init block as part of the function`, func(s *testcase.Spec) {
			entity3.Let(s, nil)

			s.Then(`it will use the var init block`, func(t *testcase.T) {
				require.True(t, entity3.Get(t).(Entity).TS == 42)
			})
		})

		s.And(`var is bound to a spec with a new let variable init block as part of the function parameter`, func(s *testcase.Spec) {
			entity3.Let(s, func(t *testcase.T) interface{} {
				return Entity{TS: 24}
			})

			s.Then(`it will use passed let init block`, func(t *testcase.T) {
				require.True(t, entity3.Get(t).(Entity).TS == 24)
			})
		})

	})
}

func TestVar_Get_nil(t *testing.T) {
	s := testcase.NewSpec(t)

	v := s.Let(`value[interface{}]`, func(t *testcase.T) interface{} {
		return nil
	})

	s.Test(``, func(t *testcase.T) {
		require.Nil(t, v.Get(t))
	})
}

func TestVar_Let_initBlock(t *testing.T) {
	s := testcase.NewSpec(t)

	type Entity struct {
		V int
	}

	s.When(`init block is absent`, func(s *testcase.Spec) {
		entity := testcase.Var{Name: "entity 1"}

		//s.And(`var is bound to a spec without providing a let variable init block as part of the function`, func(s *testcase.Spec) {
		//	require.Panics(t, func() {
		//		entity.Let(s, nil)
		//	})
		//})

		s.And(`var is bound to a spec with a new let variable init block as part of the function parameter`, func(s *testcase.Spec) {
			entity.Let(s, func(t *testcase.T) interface{} {
				return Entity{V: 42}
			})

			s.Then(`it will use passed let init block`, func(t *testcase.T) {
				require.True(t, entity.Get(t).(Entity).V == 42)
			})
		})
	})

	s.When(`init block defined for the variable`, func(s *testcase.Spec) {
		entity := testcase.Var{
			Name: "entity 2",
			Init: func(t *testcase.T) interface{} {
				return Entity{V: 84}
			},
		}

		s.And(`var is bound to a spec without providing a let variable init block as part of the function`, func(s *testcase.Spec) {
			entity.Let(s, nil)

			s.Then(`it will use the var init block`, func(t *testcase.T) {
				require.True(t, entity.Get(t).(Entity).V == 84)
			})
		})

		s.And(`var is bound to a spec with a new let variable init block as part of the function parameter`, func(s *testcase.Spec) {
			entity.Let(s, func(t *testcase.T) interface{} {
				return Entity{V: 168}
			})

			s.Then(`it will use passed let init block`, func(t *testcase.T) {
				require.True(t, entity.Get(t).(Entity).V == 168)
			})
		})

	})

	s.When(`init block defined through Spec#Let`, func(s *testcase.Spec) {
		entity := s.Let(`entity 3`, func(t *testcase.T) interface{} {
			return Entity{V: 336}
		})

		s.Test(``, func(t *testcase.T) {
			require.NotNil(t, entity.Init)
			require.True(t, 336 == entity.Init(t).(Entity).V)
		})
	})
}

func TestSpec_LetValue_returnsVar(t *testing.T) {
	s := testcase.NewSpec(t)

	const varName = `counter`
	counter := s.LetValue(varName, 0)

	s.Test(``, func(t *testcase.T) {
		require.Equal(t, 0, counter.Get(t).(int))
		t.Let(varName, 1)
		require.Equal(t, 1, counter.Get(t).(int))
		counter.Set(t, 2)
		require.Equal(t, 2, counter.Get(t).(int))
		require.Equal(t, 2, t.I(varName).(int))
	})
}

func TestVar_EagerLoading_daisyChain(t *testing.T) {
	s := testcase.NewSpec(t)

	value := s.Let(`eager loading value`, func(t *testcase.T) interface{} {
		return 42
	}).EagerLoading(s)

	s.Test(`EagerLoading returns the var object for syntax sugar purposes`, func(t *testcase.T) {
		require.Equal(t, 42, value.Get(t).(int))
	})
}

func TestAppend(t *testing.T) {
	s := testcase.NewSpec(t)

	var (
		v       = testcase.Var{Name: `testcase.Var`}
		e       = testcase.Var{Name: `new slice element`}
		subject = func(t *testcase.T) {
			testcase.Append(t, v, e.Get(t))
		}
	)

	s.When(`var content is a slice[T]`, func(s *testcase.Spec) {
		v.Let(s, func(t *testcase.T) interface{} {
			return []int{}
		})

		s.And(`the element is a T type`, func(s *testcase.Spec) {
			e.Let(s, func(t *testcase.T) interface{} {
				return fixtures.Random.Int()
			})

			s.Then(`it will append the value to the slice[T] type testcase.Var`, func(t *testcase.T) {
				require.Len(t, v.Get(t).([]int), 0)
				subject(t)

				list := v.Get(t).([]int)
				elem := e.Get(t).(int)
				require.Len(t, list, 1)
				require.Contains(t, list, elem)
			})

			s.Then(`on multiple use it will append all`, func(t *testcase.T) {
				var expected []int
				for i := 0; i < 1024; i++ {
					expected = append(expected, i)
					e.Set(t, i)
					subject(t)
				}

				require.Equal(t, expected, v.Get(t).([]int))
			})
		})
	})

	s.Test(`multiple value`, func(t *testcase.T) {
		listVar := testcase.Var{Name: `slice[T]`, Init: func(t *testcase.T) interface{} { return []string{} }}
		testcase.Append(t, listVar, `foo`, `bar`, `baz`)

		require.Equal(t, []string{`foo`, `bar`, `baz`}, listVar.Get(t).([]string))
	})
}

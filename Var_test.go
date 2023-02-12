package testcase_test

import (
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/adamluzsi/testcase/assert"
	"github.com/adamluzsi/testcase/internal/doubles"
	"github.com/adamluzsi/testcase/random"

	"github.com/adamluzsi/testcase"
)

func TestVar(t *testing.T) {
	s := testcase.NewSpec(t)

	// to testCase testVar, I need side effect by resetting the expected in a before hook
	// the var of testVar needs to be leaked into the testing subjects,
	// and I can't use a testVar to testCase testVar because I need this expected at spec level as well.
	// So to testCase testcase.Var, I can't use fully testcase.Var.
	// This should not be the case for anything else outside the testing framework.
	s.HasSideEffect()
	rnd := random.New(random.CryptoSeed{})
	var testVar = testcase.Var[int]{ID: rnd.StringNWithCharset(5, "abcdefghijklmnopqrstuvwxyz")}
	expected := rnd.Int()

	stub := &doubles.TB{}
	willFatal := willFatalWithMessageFn(stub)
	willFatalWithVariableNotFoundMessage := func(s *testcase.Spec, tb testing.TB, varName string, blk func(*testcase.T)) {
		tct := testcase.NewT(stub, s)
		assert.Must(tb).Contain(willFatal(t, func() { blk(tct) }),
			fmt.Sprintf("Variable %q is not found.", varName))
	}

	s.Describe(`#Get`, func(s *testcase.Spec) {
		subject := func(t *testcase.T) int {
			return testVar.Get(t)
		}

		s.When(`no expected defined in the spec and no init logic provided`, func(s *testcase.Spec) {
			s.Then(`it will panic, and warn about the unknown expected`, func(t *testcase.T) {
				willFatalWithVariableNotFoundMessage(s, t, testVar.ID, func(t *testcase.T) { subject(t) })
			})
		})

		s.When(`spec has value by testCase runtime Var#Set`, func(s *testcase.Spec) {
			s.Before(func(t *testcase.T) {
				testVar.Set(t, expected)
			})

			s.Then(`the expected is returned`, func(t *testcase.T) {
				assert.Must(t).Equal(expected, testVar.Get(t))
			})
		})

		s.When(`spec has value set by Var#Let`, func(s *testcase.Spec) {
			testVar.Let(s, func(t *testcase.T) int {
				return expected
			})

			s.Then(`the expected is returned`, func(t *testcase.T) {
				assert.Must(t).Equal(expected, testVar.Get(t))
			})
		})

		s.When(`spec has value set by Var#LetValue`, func(s *testcase.Spec) {
			testVar.LetValue(s, expected)

			s.Then(`the expected is returned`, func(t *testcase.T) {
				assert.Must(t).Equal(expected, testVar.Get(t))
			})
		})

		s.When(`Var#Init is defined`, func(s *testcase.Spec) {
			s.HasSideEffect()
			// WARN: do not use any other hook that manipulates the testVar here
			// else the side effect is not guaranteed
			s.Around(func(t *testcase.T) func() {
				testVar.Init = func(t *testcase.T) int { return expected }
				// reset side effect
				return func() { testVar.Init = nil }
			})

			thenValueIsCached := func(s *testcase.Spec) {
				s.Then(`value is cached`, func(t *testcase.T) {
					values := make(map[int]struct{})
					for i := 0; i < 128; i++ {
						values[testVar.Get(t)] = struct{}{}
					}
					const failReason = `it was expected that the value from the var is deterministic/cached within the testCase lifetime`
					assert.Must(t).True(len(values) == 1, failReason)
				})
			}

			s.And(`spec don't have value set in any way`, func(s *testcase.Spec) {
				s.Then(`it will return the Value from init`, func(t *testcase.T) {
					assert.Must(t).Equal(expected, testVar.Get(t))
				})

				s.And(`.Init creates a non deterministic value`, func(s *testcase.Spec) {
					s.HasSideEffect()
					testVar.Init = func(t *testcase.T) int { return t.Random.Int() }
					defer func() { testVar.Init = nil }()

					thenValueIsCached(s)
				})
			})

			s.And(`spec have value set in some form`, func(s *testcase.Spec) {
				testVar.LetValue(s, 42)

				s.Then(`the expected is returned`, func(t *testcase.T) {
					assert.Must(t).Equal(42, testVar.Get(t))
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
				assert.Must(t).Equal(expected, testVar.Get(t))
			})
		})

		s.When(`subject is not used`, func(s *testcase.Spec) {
			s.Then(`value will be absent`, func(t *testcase.T) {
				willFatalWithVariableNotFoundMessage(s, t, testVar.ID, func(t *testcase.T) { testVar.Get(t) })
			})
		})
	})

	s.Describe(`#Let`, func(s *testcase.Spec) {
		subject := func(s *testcase.Spec) {
			testVar.Let(s, func(t *testcase.T) int { return expected })
		}

		s.When(`subject used`, func(s *testcase.Spec) {
			subject(s)

			s.Then(`it will set value in the spec level`, func(t *testcase.T) {
				assert.Must(t).Equal(expected, testVar.Get(t))
			})
		})

		s.When(`subject is not used on a clean Spec`, func(s *testcase.Spec) {
			s.Then(`value will be absent`, func(t *testcase.T) {
				willFatalWithVariableNotFoundMessage(s, t, testVar.ID, func(t *testcase.T) { testVar.Get(t) })
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
				assert.Must(t).Equal(expected, testVar.Get(t))
			})
		})

		s.When(`subject is not used on a clean Spec`, func(s *testcase.Spec) {
			s.Then(`value will be absent`, func(t *testcase.T) {
				willFatalWithVariableNotFoundMessage(s, t, testVar.ID, func(t *testcase.T) { testVar.Get(t) })
			})
		})
	})

	s.Describe(`#EagerLoading`, func(s *testcase.Spec) {
		subject := func(s *testcase.Spec) {
			testVar.EagerLoading(s)
		}

		testVar.Let(s, func(t *testcase.T) int {
			return int(time.Now().UnixNano())
		})

		s.When(`subject used`, func(s *testcase.Spec) {
			subject(s)

			s.Then(`value will be eager loaded`, func(t *testcase.T) {
				time.Sleep(5 * time.Nanosecond)
				now := int(time.Now().UnixNano())
				t.Must.True(testVar.Get(t) < now)
			})
		})

		s.When(`subject not used`, func(s *testcase.Spec) {
			s.Then(`value will be lazy loaded`, func(t *testcase.T) {
				now := int(time.Now().UnixNano())
				t.Must.True(now < testVar.Get(t))
			})
		})
	})

	willFatalWithOnLetMissing := func(s *testcase.Spec, tb testing.TB, varName string, blk func(*testcase.T)) {
		tct := testcase.NewT(stub, s)
		assert.Must(tb).Contain(willFatal(t, func() { blk(tct) }),
			fmt.Sprintf("%s Var has Var.OnLet. You must use Var.Let, Var.LetValue to initialize it properly.", varName))
	}

	s.Describe(`#OnLet`, func(s *testcase.Spec) {
		s.When(`it is provided`, func(s *testcase.Spec) {
			v := testcase.Var[int]{
				ID: `foo`,
				OnLet: func(s *testcase.Spec, _ testcase.Var[int]) {
					s.Tag(`on-let`) // test trough side effect
				},
			}

			s.And(`variable is not bound to Spec`, func(s *testcase.Spec) {
				s.Test(`it will panic on Var.Get`, func(t *testcase.T) {
					willFatalWithOnLetMissing(s, t, v.ID, func(t *testcase.T) { v.Get(t) })
				})

				s.Test(`it will panic on Var.Set`, func(t *testcase.T) {
					willFatalWithOnLetMissing(s, t, v.ID, func(t *testcase.T) { v.Set(t, 42) })
				})
			})

			s.And(`variable is bound to Spec with Var.Let`, func(s *testcase.Spec) {
				v.Let(s, func(t *testcase.T) int { return 42 })

				s.Test(`Var.Get returns value`, func(t *testcase.T) {
					assert.Must(t).Equal(42, v.Get(t))
				})

				s.Test(`it will apply the setup in the context`, func(t *testcase.T) {
					assert.Must(t).True(t.HasTag(`on-let`))
				})
			})

			s.And(`variable is bound to Spec with Var.LetValue`, func(s *testcase.Spec) {
				v.LetValue(s, 42)

				s.Test(`Var.Get returns value`, func(t *testcase.T) {
					assert.Must(t).Equal(42, v.Get(t))
				})

				s.Test(`it will apply the setup in the context`, func(t *testcase.T) {
					assert.Must(t).True(t.HasTag(`on-let`))
				})
			})
		})

		s.When(`it is absent`, func(s *testcase.Spec) {
			v := testcase.Var[int]{
				ID: `foo`,
			}

			s.And(`variable is not bound to Spec`, func(s *testcase.Spec) {
				v := testcase.Var[int]{
					ID: `foo`,
					Init: func(t *testcase.T) int {
						// required to be used without binding Var to Spec
						return 42
					},
				}

				s.Test(`it will return initialized value on Var.Get`, func(t *testcase.T) {
					assert.Must(t).Equal(42, v.Get(t))
				})
			})

			s.And(`variable is bound to Spec with Var.Let`, func(s *testcase.Spec) {
				v.Let(s, func(t *testcase.T) int { return 42 })

				s.Test(`Var.Get returns value`, func(t *testcase.T) {
					assert.Must(t).Equal(42, v.Get(t))
				})

				s.Test(`no hook, no setup`, func(t *testcase.T) {
					assert.Must(t).True(!t.HasTag(`on-let`))
				})
			})

			s.And(`variable is bound to Spec with Var.LetValue`, func(s *testcase.Spec) {
				v.LetValue(s, 42)

				s.Test(`Var.Get returns value`, func(t *testcase.T) {
					assert.Must(t).Equal(42, v.Get(t))
				})

				s.Test(`no hook, no setup`, func(t *testcase.T) {
					assert.Must(t).True(!t.HasTag(`on-let`))
				})
			})
		})
	})

	s.Describe("#Super", func(s *testcase.Spec) {
		s.Test("when no super is actually set", func(t *testcase.T) {
			s := testcase.NewSpec(t)
			var v = testcase.Var[int]{ID: "Var[int]"}
			v.Let(s, func(t *testcase.T) int {
				t.Must.Panic(func() { _ = v.Super(t) })
				return 42
			})
			s.Test("", func(t *testcase.T) {
				v.Get(t)
			})
			s.Finish()
		})

		s.Test("when super is only defined with Init block", func(t *testcase.T) {
			s := testcase.NewSpec(t.TB)
			var v = testcase.Var[int]{ID: "Var[int]", Init: func(t *testcase.T) int {
				return 32
			}}
			v.Let(s, func(t *testcase.T) int {
				return v.Super(t) + 10
			})
			s.Test("", func(t *testcase.T) {
				t.Must.Equal(42, v.Get(t))
			})
		})

		s.Test("when super is defined with a testcase.Let", func(t *testcase.T) {
			s := testcase.NewSpec(t.TB)

			v := testcase.Let(s, func(t *testcase.T) int {
				return 32
			})

			s.Context("", func(s *testcase.Spec) {
				v.Let(s, func(t *testcase.T) int {
					return v.Super(t) + 10
				})

				s.Test("", func(t *testcase.T) {
					t.Must.Equal(42, v.Get(t))
				})
			})
		})

		s.Test("when super is defined with a Var[V].Let along with an Init", func(t *testcase.T) {
			s := testcase.NewSpec(t.TB)

			v := testcase.Var[int]{
				ID:   "Var[int]",
				Init: func(t *testcase.T) int { return 128 },
			}

			v.Let(s, func(t *testcase.T) int {
				return 32
			})

			s.Context("", func(s *testcase.Spec) {
				v.Let(s, func(t *testcase.T) int {
					t.Log("n-lvl-1", v.Super(t))
					return v.Super(t) + 10
				})

				s.Test("", func(t *testcase.T) {
					t.Must.Equal(42, v.Get(t))
				})
			})
		})

		s.Test("when Init is set along with OnLet, but it was never set with a Let", func(t *testcase.T) {
			s := testcase.NewSpec(t.TB)

			v := testcase.Var[int]{
				ID:    "Var[int]",
				Init:  func(t *testcase.T) int { return 32 },
				OnLet: func(s *testcase.Spec, v testcase.Var[int]) {},
			}

			v.Let(s, func(t *testcase.T) int {
				return v.Super(t) + 10
			})

			s.Test("", func(t *testcase.T) {
				t.Must.Equal(42, v.Get(t))
			})
		})

		s.Test("when super is defined with a testcase.Let", func(t *testcase.T) {
			s := testcase.NewSpec(t.TB)

			v := testcase.Let(s, func(t *testcase.T) int {
				return 32
			})

			s.Context("", func(s *testcase.Spec) {
				v.Let(s, func(t *testcase.T) int {
					return v.Super(t) + 10
				})

				s.Test("", func(t *testcase.T) {
					t.Must.Equal(42, v.Get(t))
				})
			})
		})

		s.Test("the results of super is cached", func(t *testcase.T) {
			s := testcase.NewSpec(t.TB)

			v := testcase.Let(s, func(t *testcase.T) int {
				return t.Random.Int()
			})

			s.Context("", func(s *testcase.Spec) {
				v.Let(s, func(t *testcase.T) int {
					t.Must.Equal(v.Super(t), v.Super(t))
					return 42
				})

				s.Test("", func(t *testcase.T) {
					t.Must.Equal(42, v.Get(t))
				})
			})
		})

		s.Test("when super is defined with a testcase.Let", func(t *testcase.T) {
			s := testcase.NewSpec(t.TB)

			v := testcase.Let(s, func(t *testcase.T) int {
				return t.Random.Int() + 1
			})

			s.Context("", func(s *testcase.Spec) {
				v.Let(s, func(t *testcase.T) int {
					t.Must.Equal(v.Super(t), v.Super(t))
					return v.Super(t)
				})

				s.Test("", func(t *testcase.T) {
					t.Must.NotEmpty(v.Get(t))
				})
			})
		})

		s.Test("super is inherited and always represent the previous declaration", func(t *testcase.T) {
			s := testcase.NewSpec(t.TB)

			v := testcase.Var[int]{
				ID: "v",
				Init: func(t *testcase.T) int {
					t.Log("init", 1)
					return 1
				},
			}

			v.Let(s, func(t *testcase.T) int {
				super := v.Super(t)
				t.Log("nl-0", super)
				t.Must.Equal(super, v.Super(t))
				return super + 1
			})

			s.Context("nl-1", func(s *testcase.Spec) {
				v.Let(s, func(t *testcase.T) int {
					super := v.Super(t)
					t.Log("nl-1", super)
					t.Must.Equal(super, v.Super(t))
					return super + 1
				})

				s.Then("", func(t *testcase.T) {
					t.Must.Equal(3, v.Get(t))
				})

				s.Context("nl-2", func(s *testcase.Spec) {
					v.Let(s, func(t *testcase.T) int {
						super := v.Super(t)
						t.Log("nl-2", super)
						t.Must.Equal(super, v.Super(t))
						return super + 1
					})

					s.Then("", func(t *testcase.T) {
						t.Must.Equal(4, v.Get(t))
					})

					s.Context("nl-3", func(s *testcase.Spec) {
						v.Let(s, func(t *testcase.T) int {
							super := v.Super(t)
							t.Log("nl-3", super)
							t.Must.Equal(super, v.Super(t))
							return super + 1
						})

						s.Then("", func(t *testcase.T) {
							t.Must.Equal(5, v.Get(t))
						})
					})
				})
			})
		})

		s.Test("super declaration is inherited along the nested testing context branch", func(t *testcase.T) {
			t.Log("and always the previous super call will be inherited")
			s := testcase.NewSpec(t.TB)

			v := testcase.Var[int]{
				ID: "v",
				Init: func(t *testcase.T) int {
					t.Log("n-lvl-nan -> init")
					return 32
				},
			}

			v.Let(s, func(t *testcase.T) int {
				t.Log("n-lvl-0", v.Super(t))
				return v.Super(t) + 10
			})

			s.Context("", func(s *testcase.Spec) {
				s.Context("", func(s *testcase.Spec) {
					v.Let(s, func(t *testcase.T) int {
						t.Log("n-lvl-2", v.Super(t))
						return v.Super(t) + 22
					})
					s.Context("", func(s *testcase.Spec) {
						s.Context("", func(s *testcase.Spec) {
							s.Then("", func(t *testcase.T) {
								t.Must.Equal(64, v.Get(t))
							})
						})
					})
				})
				s.Then("", func(t *testcase.T) {
					t.Must.Equal(42, v.Get(t))
				})
			})
		})

		s.Test("Var.Super can be used from an another testcase.Let as well", func(t *testcase.T) {
			t.Log("this is ideal to make a stub that wraps the original value of an another variable")
			t.Log("and then the orignal var use the result of this new variable.")
			t.Log("For example a Role Interface value is decorated with fault injection.")
			s := testcase.NewSpec(t.TB)

			v1 := testcase.Let(s, func(t *testcase.T) int {
				return 32
			})

			s.Context("", func(s *testcase.Spec) {
				v2 := testcase.Let(s, func(t *testcase.T) int {
					return v1.Super(t) + 10
				})
				v1.Let(s, func(t *testcase.T) int { return v2.Get(t) })

				s.Test("", func(t *testcase.T) {
					t.Must.Equal(42, v1.Get(t))
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

	entity1 := testcase.Let(s, func(t *testcase.T) Entity {
		return Entity{TS: time.Now().UnixNano()}
	})

	entity2 := testcase.Let(s, func(t *testcase.T) Entity {
		return Entity{TS: time.Now().UnixNano()}
	})

	s.When(`var is allowed to use lazy loading`, func(s *testcase.Spec) {
		// nothing to do here, lazy loading is the default behavior

		s.Then(`it should be initialized when it is first accessed`, func(t *testcase.T) {
			e1ts := entity1.Get(t).TS
			time.Sleep(42 * time.Nanosecond)
			e2ts := entity2.Get(t).TS
			assert.Must(t).True(e1ts < e2ts)
		})
	})

	s.When(`var eager loading is requested`, func(s *testcase.Spec) {
		entity2.EagerLoading(s)

		s.Then(`the value should be evaluated `, func(t *testcase.T) {
			e1ts := entity1.Get(t).TS
			time.Sleep(42 * time.Nanosecond)
			t.Log(`now we access entity 2,`)
			t.Log(`but the value should already be evaluated by the time the test case block is reached`)
			e2ts := entity2.Get(t).TS
			assert.Must(t).True(e2ts < e1ts)
		})
	})

	s.When(`var override done at spec spec level`, func(s *testcase.Spec) {
		entity1.Let(s, func(t *testcase.T) Entity {
			return Entity{TS: 0}
		})

		s.Then(`in the test case the overridden value will be the initial value`, func(t *testcase.T) {
			assert.Must(t).True(entity1.Get(t).TS == 0)
		})

		s.Context(``, func(s *testcase.Spec) {
			entity1.Let(s, func(t *testcase.T) Entity {
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
			assert.Must(t).True(entity1.Get(t).TS == 0)
		})
	})

	s.Context(`var override at testCase runtime level`, func(s *testcase.Spec) {
		s.Before(func(t *testcase.T) {
			entity1.Set(t, Entity{TS: 0})
		})

		s.Then(``, func(t *testcase.T) {
			assert.Must(t).True(entity1.Get(t).TS == 0)
		})
	})

	s.When(`init block defined for the variable`, func(s *testcase.Spec) {
		entity3 := testcase.Var[Entity]{
			ID: "entity 4",
			Init: func(t *testcase.T) Entity {
				return Entity{TS: 42}
			},
		}

		s.And(`var is bound to a spec without providing a let variable init block as part of the function`, func(s *testcase.Spec) {
			entity3.Let(s, nil)

			s.Then(`it will use the var init block`, func(t *testcase.T) {
				assert.Must(t).True(entity3.Get(t).TS == 42)
			})
		})

		s.And(`var is bound to a spec with a new let variable init block as part of the function parameter`, func(s *testcase.Spec) {
			entity3.Let(s, func(t *testcase.T) Entity {
				return Entity{TS: 24}
			})

			s.Then(`it will use passed let init block`, func(t *testcase.T) {
				assert.Must(t).True(entity3.Get(t).TS == 24)
			})
		})

	})
}

func TestVar_Get_interface_as_nil(t *testing.T) {
	s := testcase.NewSpec(t)

	v := testcase.Let(s, func(t *testcase.T) interface{} {
		return nil
	})

	s.Test(``, func(t *testcase.T) {
		t.Must.Nil(v.Get(t))
	})
}
func TestVar_Get_pointer_as_nil(t *testing.T) {
	s := testcase.NewSpec(t)

	type T struct{}

	v := testcase.Let(s, func(t *testcase.T) *T {
		return nil
	})

	s.Test(``, func(t *testcase.T) {
		t.Must.Nil(v.Get(t))
	})
}

func TestVar_Get_threadSafe(t *testing.T) {
	s := testcase.NewSpec(t)
	v := testcase.Var[int]{
		ID:    `num`,
		Init:  func(t *testcase.T) int { return 0 },
		OnLet: func(s *testcase.Spec, _ testcase.Var[int]) {},
	}
	v.Let(s, nil)
	s.Test(``, func(t *testcase.T) {
		blk := func() {
			value := v.Get(t)
			v.Set(t, value+1)
		}

		testcase.Race(blk, blk)
	})
}

func TestVar_Get_valueSetDuringAnotherVarInitBlock(t *testing.T) {
	unsupported(t)

	s := testcase.NewSpec(t)

	getValues := func(t *testcase.T) (int, int) {
		return t.Random.Int(), t.Random.Int()
	}

	var a, b testcase.Var[int]

	a = testcase.Var[int]{
		ID: `A`,
		Init: func(t *testcase.T) int {
			av, bv := getValues(t)
			b.Set(t, bv)
			return av
		},
	}
	b = testcase.Var[int]{
		ID: `B`,
		Init: func(t *testcase.T) int {
			a.Get(t) // lazy load init
			return b.Get(t)
		},
	}

	s.Test(``, func(t *testcase.T) {
		b.Get(t)
	})
}

func TestVar_Let_initBlock(t *testing.T) {
	s := testcase.NewSpec(t)

	type Entity struct {
		V int
	}

	s.When(`init block is absent`, func(s *testcase.Spec) {
		entity := testcase.Var[Entity]{ID: "entity 1"}

		//s.And(`var is bound to a spec without providing a Let variable init block as part of the function`, func(s *testcase.Spec) {
		//	assert.Must(t).Panic(func() {
		//		entity.Let(s, nil)
		//	})
		//})

		s.And(`var is bound to a spec with a new Let variable init block as part of the function parameter`, func(s *testcase.Spec) {
			entity.Let(s, func(t *testcase.T) Entity {
				return Entity{V: 42}
			})

			s.Then(`it will use passed Let init block`, func(t *testcase.T) {
				assert.Must(t).True(entity.Get(t).V == 42)
			})
		})
	})

	s.When(`init block defined for the variable`, func(s *testcase.Spec) {
		entity := testcase.Var[Entity]{
			ID: "entity 2",
			Init: func(t *testcase.T) Entity {
				return Entity{V: 84}
			},
		}

		s.And(`var is bound to a spec without providing a Let variable init block as part of the function`, func(s *testcase.Spec) {
			entity.Let(s, nil)

			s.Then(`it will use the var init block`, func(t *testcase.T) {
				assert.Must(t).True(entity.Get(t).V == 84)
			})
		})

		s.And(`var is bound to a spec with a new Let variable init block as part of the function parameter`, func(s *testcase.Spec) {
			entity.Let(s, func(t *testcase.T) Entity {
				return Entity{V: 168}
			})

			s.Then(`it will use passed Let init block`, func(t *testcase.T) {
				assert.Must(t).True(entity.Get(t).V == 168)
			})
		})

	})

	s.When(`init block defined through Spec#Let`, func(s *testcase.Spec) {
		entity := testcase.Let(s, func(t *testcase.T) interface{} {
			return Entity{V: 336}
		})

		s.Test(``, func(t *testcase.T) {
			t.Must.NotNil(entity.Init)
			t.Must.True(336 == entity.Init(t).(Entity).V)
		})
	})
}

func TestVar_EagerLoading_daisyChain(t *testing.T) {
	s := testcase.NewSpec(t)

	value := testcase.Let(s, func(t *testcase.T) interface{} {
		return 42
	}).EagerLoading(s)

	s.Test(`EagerLoading returns the var object for syntax sugar purposes`, func(t *testcase.T) {
		assert.Must(t).Equal(42, value.Get(t))
	})
}

func TestAppend(t *testing.T) {
	s := testcase.NewSpec(t)

	var (
		v       = testcase.Var[any]{ID: `testcase.Var`}
		e       = testcase.Var[any]{ID: `new slice element`}
		subject = func(t *testcase.T) {
			testcase.Append(t, v, e.Get(t))
		}
	)

	s.When(`var content is a slice[T]`, func(s *testcase.Spec) {
		v.Let(s, func(t *testcase.T) any {
			return []int{}
		})

		s.And(`the element is a T type`, func(s *testcase.Spec) {
			e.Let(s, func(t *testcase.T) interface{} {
				return t.Random.Int()
			})

			s.Then(`it will append the value to the slice[T] type testcase.Var`, func(t *testcase.T) {
				t.Must.Equal(len(v.Get(t).([]int)), 0)
				subject(t)

				list := v.Get(t)
				elem := e.Get(t)
				t.Must.Equal(len(list.([]int)), 1)
				t.Must.Contain(list, elem)
			})

			s.Then(`on multiple use it will append all`, func(t *testcase.T) {
				var expected []int
				for i := 0; i < 1024; i++ {
					expected = append(expected, i)
					e.Set(t, i)
					subject(t)
				}

				assert.Must(t).Equal(expected, v.Get(t))
			})
		})
	})

	s.Test(`multiple value`, func(t *testcase.T) {
		listVar := testcase.Var[[]string]{ID: `slice[T]`, Init: func(t *testcase.T) []string { return []string{} }}
		testcase.Append(t, listVar, `foo`, `bar`, `baz`)
		assert.Must(t).Equal([]string{`foo`, `bar`, `baz`}, listVar.Get(t))
	})
}

func TestVar_Get_concurrentInit_initOnlyOnce(t *testing.T) {
	s := testcase.NewSpec(t)
	var (
		mutex    sync.Mutex
		counter  int
		variable = testcase.Let(s, func(t *testcase.T) interface{} {
			mutex.Lock()
			counter++
			mutex.Unlock()
			return t.Random.Int()
		})
	)
	s.Test(``, func(t *testcase.T) {
		blk := func() { _ = variable.Get(t) }
		var blks []func()
		for i := 0; i < 42; i++ {
			blks = append(blks, blk)
		}
		testcase.Race(blk, blk, blks...)
		assert.Must(t).Equal(1, counter)
	})
}

func TestVar_Get_race(t *testing.T) {
	var (
		s       = testcase.NewSpec(t)
		a       = testcase.Let(s, func(t *testcase.T) int { return t.Random.Int() })
		b       = testcase.Let(s, func(t *testcase.T) int { return t.Random.Int() })
		c       = testcase.Let(s, func(t *testcase.T) int { return b.Get(t) })
		subject = func(t *testcase.T) int { return a.Get(t) + c.Get(t) }
	)
	s.Test(``, func(t *testcase.T) {
		blk := func() { _ = subject(t) }
		testcase.Race(blk, blk, blk)
	})
}

func TestVar_Bind(t *testing.T) {
	rnd := random.New(random.CryptoSeed{})
	t.Run("bind eill ensure that the variable is bound to the specification", func(t *testing.T) {
		s := testcase.NewSpec(t)
		expected := rnd.Int()
		var onLetRan bool
		v := testcase.Var[int]{
			ID: "variable", Init: func(t *testcase.T) int { return expected },
			OnLet: func(s *testcase.Spec, self testcase.Var[int]) { onLetRan = true },
		}
		v2 := v.Bind(s)
		assert.Must(t).Equal(v.ID, v2.ID)
		s.Test(``, func(t *testcase.T) {
			assert.Must(t).Equal(expected, v.Get(t))
		})
		s.Finish()
		assert.True(t, onLetRan)
	})
	t.Run("bind will not overrite a previous value assignment", func(t *testing.T) {
		s := testcase.NewSpec(t)
		expected := rnd.Int()
		var onLetRan bool
		v := testcase.Var[int]{
			ID: "variable", Init: func(t *testcase.T) int { return rnd.Int() },
			OnLet: func(s *testcase.Spec, self testcase.Var[int]) { onLetRan = true },
		}
		v.Let(s, func(s *testcase.T) int { return expected })
		v = v.Bind(s)
		s.Test(``, func(t *testcase.T) {
			assert.Must(t).Equal(expected, v.Get(t))
		})
		s.Finish()
		assert.True(t, onLetRan)
	})
}

func TestVar_Before(t *testing.T) {
	t.Run(`When var not bounded to the Spec, then it will execute on Var.Get`, func(t *testing.T) {
		s := testcase.NewSpec(t)
		executed := testcase.LetValue(s, false)
		v := testcase.Var[int]{
			ID: "variable",
			Init: func(t *testcase.T) int {
				return t.Random.Int()
			},
			Before: func(t *testcase.T, v testcase.Var[int]) { executed.Set(t, true) },
		}
		s.Test(``, func(t *testcase.T) {
			assert.Must(t).True(!executed.Get(t))
			_ = v.Get(t)
			assert.Must(t).True(executed.Get(t))
		})
	})
	t.Run(`When Var initialized by an other Var, Before can eager load the other variable on Var.Get`, func(t *testing.T) {
		expected := random.New(random.CryptoSeed{}).Int()
		var sbov, oth testcase.Var[int]
		oth = testcase.Var[int]{ID: "other variable", Init: func(t *testcase.T) int {
			sbov.Set(t, expected)
			return 42
		}}
		sbov = testcase.Var[int]{ID: "set by other variable", Before: func(t *testcase.T, _ testcase.Var[int]) {
			oth.Get(t)
		}}
		s := testcase.NewSpec(t)
		s.Test(``, func(t *testcase.T) {
			assert.Must(t).Equal(expected, sbov.Get(t))
		})
	})
	t.Run(`calling Var.Get from the .Before block should not cause an issue`, func(t *testing.T) {
		var v testcase.Var[int]
		v = testcase.Var[int]{
			ID: "variable",
			Init: func(t *testcase.T) int {
				return 42
			},
			Before: func(t *testcase.T, v testcase.Var[int]) {
				t.Logf("v value: %v", v.Get(t))
			},
		}
		s := testcase.NewSpec(t)
		s.Test(``, func(t *testcase.T) {
			_ = v.Get(t)
		})
	})
	t.Run(`when Var bound to the Spec.Context, before is executed early on`, func(t *testing.T) {
		s := testcase.NewSpec(t)

		executed := testcase.LetValue(s, false)
		v := testcase.Var[int]{
			ID: "variable",
			Init: func(t *testcase.T) int {
				return t.Random.Int()
			},
			Before: func(t *testcase.T, v testcase.Var[int]) { executed.Set(t, true) },
		}

		v.Bind(s)

		s.Test(``, func(t *testcase.T) {
			t.Must.True(executed.Get(t))
			_ = v.Get(t)
			t.Must.True(executed.Get(t))
		})
	})
	t.Run(`The Before func receive the value of the variable as second argument`, func(t *testing.T) {
		s := testcase.NewSpec(t)

		vFromBefore := testcase.LetValue[int](s, 0)
		v := testcase.Var[int]{
			ID: "variable",
			Init: func(t *testcase.T) int {
				return t.Random.Int()
			},
			Before: func(t *testcase.T, v testcase.Var[int]) {
				vFromBefore.Set(t, v.Get(t))
			},
		}

		v.Bind(s)

		s.Test(``, func(t *testcase.T) {
			expected := v.Get(t)
			actual := vFromBefore.Get(t)
			t.Must.Equal(expected, actual)
		})
	})
}

func TestVar_missingID(t *testing.T) {
	varWithoutID := testcase.Var[string]{}
	stub := &doubles.TB{}
	tct := testcase.NewT(stub, nil)
	assert.Panic(t, func() { _ = varWithoutID.Get(tct) })
	assert.Contain(t, stub.Logs.String(), "ID for testcase.Var[string] is missing. Maybe it's uninitialized?")
}

func TestVar_PreviousValue_smoke(t *testing.T) {
	s := testcase.NewSpec(t)

	v := testcase.Let(s, func(t *testcase.T) int {
		return 32
	})

	s.Context("", func(s *testcase.Spec) {
		v.Let(s, func(t *testcase.T) int {
			return v.PreviousValue(t) + 10
		})

		s.Test("", func(t *testcase.T) {
			t.Must.Equal(42, v.Get(t))
		})
	})
}

//func TestCastToVarInit(t *testing.T) {
//	s := testcase.NewSpec(t)
//
//	var (
//		fn = testcase.Let[func(testing.TB) int](s, nil)
//	)
//	act := func(t *testcase.T) func(*testcase.T) int {
//		return testcase.CastToVarInit(fn.Get(t))
//	}
//
//	s.When("fn is nil", func(s *testcase.Spec) {
//		fn.LetValue(s, nil)
//
//		s.Then("nil is returned", func(t *testcase.T) {
//			t.Must.Nil(act(t))
//		})
//	})
//
//	s.When("fn is valid", func(s *testcase.Spec) {
//		expectedValue := testcase.Let(s, func(t *testcase.T) int {
//			return t.Random.Int()
//		})
//
//		fn.Let(s, func(t *testcase.T) func(testing.TB) int {
//			return func(tb testing.TB) int {
//				t.Must.NotNil(tb)
//				return expectedValue.Get(t)
//			}
//		})
//
//		s.Then("init function is used", func(t *testcase.T) {
//			varInit := act(t)
//			t.Must.NotNil(varInit)
//			t.Must.Equal(expectedValue.Get(t), varInit(t))
//		})
//	})
//
//	s.Context("smoke", func(s *testcase.Spec) {
//		makeFn := func(testing.TB) string {
//			return "The answer"
//		}
//		v := testcase.Let(s, testcase.CastToVarInit(makeFn))
//
//		s.Then("", func(t *testcase.T) {
//			t.Must.Equal("The answer", v.Get(t))
//		})
//	})
//}

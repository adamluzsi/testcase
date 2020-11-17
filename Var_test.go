package testcase_test

import (
	"github.com/adamluzsi/testcase"
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

func TestVar(t *testing.T) {
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
			t.Log(`but the value should already be evaluated by the time the test case block is reached`)
			e2ts := entity2.Get(t).(Entity).TS
			require.True(t, e2ts < e1ts)
		})
	})

	s.When(`var override done at spec context level`, func(s *testcase.Spec) {
		entity1.Let(s, func(t *testcase.T) interface{} {
			return Entity{TS: 0}
		})

		s.Then(`in the test case the overridden value will be the initial value`, func(t *testcase.T) {
			require.True(t, entity1.Get(t).(Entity).TS == 0)
		})

		s.Context(``, func(s *testcase.Spec) {
			entity1.Let(s, func(t *testcase.T) interface{} {
				// defined at spec level -> will be initial value with lazy load
				return Entity{TS: time.Now().UnixNano()}
			})
			s.Before(func(t *testcase.T) {
				// defined at test run time, will be eager loaded
				entity2.Set(t, Entity{TS: time.Now().UnixNano()})
			})

			s.Test(`spec level definition should be the `, func(t *testcase.T) {

			})
		})
	})

	s.When(`var override done at test runtime level`, func(s *testcase.Spec) {
		s.Before(func(t *testcase.T) {
			entity1.Set(t, Entity{TS: 0})
		})

		s.Then(``, func(t *testcase.T) {
			require.True(t, entity1.Get(t).(Entity).TS == 0)
		})
	})

	s.Context(`var override at test runtime level`, func(s *testcase.Spec) {
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

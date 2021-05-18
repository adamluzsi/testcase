package testcase_test

import (
	"context"
	"testing"

	"github.com/adamluzsi/testcase"
)

func ExampleT_Set() {
	var t *testing.T
	var s = testcase.NewSpec(t)
	s.Parallel()

	ctx := s.Let(`ctx`, func(t *testcase.T) interface{} {
		return context.Background()
	})

	s.When(`let can be manipulated during runtime hooks by simply calling *T#Let`, func(s *testcase.Spec) {
		s.Before(func(t *testcase.T) {
			newContext := context.WithValue(ctx.Get(t).(context.Context), `certain`, `value`)

			// here for example we update the testCase variable ctx to have a certain value to fulfil the subcontext goal
			t.Set(ctx.Name, newContext)
			// or with variable setter
			ctx.Set(t, newContext)
		})

		s.Then(`ctx here has the value that was assigned in the before hook`, func(t *testcase.T) {
			_ = ctx.Get(t).(context.Context)
		})
	})

	s.Then(`your ctx is in the original state without any modifications`, func(t *testcase.T) {
		_ = ctx.Get(t).(context.Context)
	})
}

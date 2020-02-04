package testcase_test

import (
	"context"
	"testing"

	"github.com/adamluzsi/testcase"
)

func ExampleT_Let(t *testing.T) {
	var s = testcase.NewSpec(t)
	s.Parallel()

	s.Let(`ctx`, func(t *testcase.T) interface{} {
		return context.Background()
	})

	s.When(`let can be manipulated during runtime hooks by simply calling *T#Let`, func(s *testcase.Spec) {
		s.Before(func(t *testcase.T) {
			t.Log(`here for example we update the test varibale ctx to have a certain value to fulfil the subcontext goal`)
			t.Let(`ctx`, context.WithValue(t.I(`ctx`).(context.Context), `certain`, `value`))
		})

		s.Then(`ctx here has the value that was assigned in the before hook`, func(t *testcase.T) {
			_ = t.I(`ctx`).(context.Context)
		})
	})

	s.Then(`your ctx is in the original state without any modifications`, func(t *testcase.T) {
		_ = t.I(`ctx`).(context.Context)
	})
}

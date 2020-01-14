package testcase_test

import (
	"context"
	"testing"

	"github.com/adamluzsi/testcase"
)

func ExampleT_Let(t *testing.T) {
	var s = testcase.NewSpec(t)
	var myLetTestExampleSubjectFunction = func(ctx context.Context) {}

	s.Describe(`my func`, func(s *testcase.Spec) {
		var subject = func(t *testcase.T) {
			myLetTestExampleSubjectFunction(t.I(`ctx`).(context.Context))
		}

		s.Let(`ctx`, func(t *testcase.T) interface{} {
			return context.Background()
		})

		s.When(`context has a certain value`, func(s *testcase.Spec) {
			s.Before(func(t *testcase.T) {
				// of course you can use oneline as well
				currentCtxValue := t.I(`ctx`).(context.Context)
				nextCtxValue := context.WithValue(currentCtxValue, `certain`, `value`)
				t.Let(`ctx`, nextCtxValue)
			})

			s.Then(`your subject here will receive a ctx that has the value`, func(t *testcase.T) {
				subject(t)
			})
		})

		s.Then(`your subject here will receive a ctx that is unaffected by the subcontext context manipulation`, func(t *testcase.T) {
			subject(t)
		})
	})
}

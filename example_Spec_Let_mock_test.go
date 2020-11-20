//go:generate mockgen -source example_Spec_Let_mock_test.go -destination example_Spec_Let_mock_mocks_test.go -package testcase_test
package testcase_test

import (
	"testing"

	"github.com/golang/mock/gomock"

	"github.com/adamluzsi/testcase"
)

func ExampleSpec_Let_mock() {
	var t *testing.T
	s := testcase.NewSpec(t)

	mock := s.Let(`the-mock`, func(t *testcase.T) interface{} {
		ctrl := gomock.NewController(t)
		mock := NewMockInterfaceExample(ctrl)
		t.Defer(ctrl.Finish)
		return mock
	})

	s.When(`some scope where mock should behave in a certain way`, func(s *testcase.Spec) {
		s.Before(func(t *testcase.T) {
			mock.Get(t).(*MockInterfaceExample).
				EXPECT().
				Say().
				Return(`some value but can also be a value from *testcase.variables`)
		})

		s.Then(`mock will be available in every test case and finish called afterwards`, func(t *testcase.T) {
			// ...
		})
	})
}

type InterfaceExample interface {
	Say() string
}

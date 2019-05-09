package testcase_test

import (
	"github.com/golang/mock/gomock"
	"testing"

	"github.com/adamluzsi/testcase"
)

func ExampleSpec_Let_howToUseMocks(t *testing.T) {
	s := testcase.NewSpec(t)

	s.Let(`the-mock-ctrl`, func(t *testcase.T) interface{} {
		return gomock.NewController(t)
	})

	s.Let(`the-mock`, func(t *testcase.T) interface{} {
		return NewMockInterfaceExample(t.I(`the-mock-ctrl`).(*gomock.Controller))
	})

	s.After(func(t *testcase.T) {
		t.I(`the-mock-ctrl`).(*gomock.Controller).Finish()
	})

	s.When(`some scope where mock should behave in a certain way`, func(s *testcase.Spec) {
		s.Before(func(t *testcase.T) {
			t.I(`*MockInterfaceExample`).(*MockInterfaceExample).
				EXPECT().
				Say().
				Return(`some value but can also be a value from *testcase.V`)
		})

		s.Then(`mock will be available in every test case and finish called afterwards`, func(t *testcase.T) {
			// OK
		})
	})
}

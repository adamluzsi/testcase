//go:generate mockgen -source DependsOnFailable.go -destination DependsOnFailable_mocks_test.go -package examples_test
package examples_test

import (
	"errors"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"

	"github.com/adamluzsi/testcase"
	"github.com/adamluzsi/testcase/examples"
)

func TestDependsOnFailable_Spec(t *testing.T) {
	s := testcase.NewSpec(t)

	var (
		failable *MockFailable
	)
	s.Around(func(t *testcase.T) func() {
		ctrl := gomock.NewController(t)
		failable = NewMockFailable(ctrl)
		return ctrl.Finish
	})
	dependsOnFailable := func(t *testcase.T) *examples.DependsOnFailable {
		return &examples.DependsOnFailable{Failable: failable}
	}

	s.Describe(`Run`, func(s *testcase.Spec) {
		subject := func(t *testcase.T) error {
			return dependsOnFailable(t).Run()
		}

		s.When(`failable fails`, func(s *testcase.Spec) {
			const expectedErrorMessage = `boom`
			s.Before(func(t *testcase.T) {
				failable.EXPECT().Do().Return(errors.New(expectedErrorMessage))
			})

			s.Then(`it will propagate back the error`, func(t *testcase.T) {
				require.EqualError(t, subject(t), expectedErrorMessage)
			})
		})

		s.When(`failable succeeds`, func(s *testcase.Spec) {
			s.Before(func(t *testcase.T) { failable.EXPECT().Do().Return(nil) })

			s.Then(`it will succeed`, func(t *testcase.T) {
				require.Nil(t, subject(t))
			})
		})
	})
}

func TestDependsOnFailable_Steps(t *testing.T) {
	s := testcase.Steps{}

	var (
		failable *MockFailable
	)
	s = s.Around(func(t *testing.T) func() {
		ctrl := gomock.NewController(t)
		failable = NewMockFailable(ctrl)
		return ctrl.Finish
	})
	dependsOnFailable := func() *examples.DependsOnFailable {
		return &examples.DependsOnFailable{Failable: failable}
	}

	t.Run(`Run`, func(t *testing.T) {
		subject := func() error {
			return dependsOnFailable().Run()
		}

		t.Run(`when failable fails`, func(t *testing.T) {
			const expectedErrorMessage = `boom`
			s := s.Before(func(t *testing.T) {
				failable.EXPECT().Do().Return(errors.New(expectedErrorMessage))
			})

			t.Run(`then it will propagate back the error`, func(t *testing.T) {
				defer s.Setup(t)()

				require.EqualError(t, subject(), expectedErrorMessage)
			})
		})

		t.Log(`when failable passes`)
		s = s.Before(func(t *testing.T) { failable.EXPECT().Do().Return(nil) })

		t.Run(`then it will succeed`, func(t *testing.T) {
			defer s.Setup(t)()
			require.Nil(t, subject())
		})

	})

}

//go:generate mockgen -source DependsOnFailable.go -destination DependsOnFailable_mocks_test.go -package examples_test
package examples_test

import (
	"errors"
	"testing"

	"github.com/adamluzsi/testcase"
	"github.com/adamluzsi/testcase/docs/examples"
	"github.com/golang/mock/gomock"
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
			var expectedErr = errors.New(`boom`)
			s.Before(func(t *testcase.T) {
				failable.EXPECT().Do().Return(expectedErr)
			})

			s.Then(`it will propagate back the error`, func(t *testcase.T) {
				t.Must.Equal(subject(t), expectedErr)
			})
		})

		s.When(`failable succeeds`, func(s *testcase.Spec) {
			s.Before(func(t *testcase.T) { failable.EXPECT().Do().Return(nil) })

			s.Then(`it will succeed`, func(t *testcase.T) {
				t.Must.Nil(subject(t))
			})
		})
	})
}

package examples_test

import (
	"math/rand"
	"strings"
	"testing"

	"github.com/adamluzsi/testcase"
	"github.com/adamluzsi/testcase/assert"
	"github.com/adamluzsi/testcase/docs/examples"
)

func TestValidateName(t *testing.T) {
	s := testcase.NewSpec(t)

	name := testcase.Var[string]{ID: `name`}

	var subject = func(t *testcase.T) error {
		return examples.ValidateName(name.Get(t))
	}

	s.When(`is perfect`, func(s *testcase.Spec) {
		name.LetValue(s, `The answer is 42`)

		s.Then(`it will be accepted without a problem`, func(t *testcase.T) {
			assert.Must(t).Nil(subject(t))
		})
	})

	s.When(`is really long`, func(s *testcase.Spec) {
		name.LetValue(s, strings.Repeat(`x`, 128+rand.Intn(42)+1))

		s.Then(`it will that the name is too long`, func(t *testcase.T) {
			assert.Must(t).Equal(examples.ErrTooLong, subject(t))
		})
	})
}

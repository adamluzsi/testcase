package examples_test

import (
	"math/rand"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/adamluzsi/testcase"
	"github.com/adamluzsi/testcase/docs/examples"
)

func TestValidateName(t *testing.T) {
	s := testcase.NewSpec(t)

	var subject = func(t *testcase.T) error {
		return examples.ValidateName(t.I(`name`).(string))
	}

	s.When(`is perfect`, func(s *testcase.Spec) {
		s.LetValue(`name`, `The answer is 42`)

		s.Then(`it will be accepted without a problem`, func(t *testcase.T) {
			require.Nil(t, subject(t))
		})
	})

	s.When(`is really long`, func(s *testcase.Spec) {
		s.LetValue(`name`, strings.Repeat(`x`, 128+rand.Intn(42)+1))

		s.Then(`it will that the name is too long`, func(t *testcase.T) {
			require.Equal(t, examples.ErrTooLong, subject(t))
		})
	})
}

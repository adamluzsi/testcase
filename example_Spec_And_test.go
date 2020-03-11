package testcase_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/adamluzsi/testcase"
)

func ExampleSpec_And() {
	var t *testing.T
	s := testcase.NewSpec(t)

	myType := func(t *testcase.T) *MyType {
		return &MyType{Field1: t.I(`input`).(string)}
	}

	s.When(`input has upcase letter`, func(s *testcase.Spec) {
		s.LetValue(`input`, `UPPER`)

		s.And(`mixed with lowercase letters`, func(s *testcase.Spec) {
			s.LetValue(`input`, `UPPERlower`)

			s.Then(`it will be false`, func(t *testcase.T) {
				require.False(t, myType(t).IsLower())
			})
		})

		s.And(`input is all upcase letter`, func(s *testcase.Spec) {
			s.Then(`it will be false`, func(t *testcase.T) {
				require.False(t, myType(t).IsLower())
			})
		})

		s.Then(`it will be false`, func(t *testcase.T) {
			require.False(t, myType(t).IsLower())
		})
	})
}

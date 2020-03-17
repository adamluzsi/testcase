package httpspec

import (
	"github.com/adamluzsi/testcase"
)

func GivenThisIsAnAPI(s *testcase.Spec) {
	setup(s)
}

func GivenThisIsAJSONAPI(s *testcase.Spec) {
	setup(s)

	s.Before(func(t *testcase.T) {
		Header(t).Set(`Content-Type`, `application/json`)
	})
}

package httpspec

import (
	"github.com/adamluzsi/testcase"
)

func ContentTypeIsJSON(s *testcase.Spec) {
	s.Before(func(t *testcase.T) {
		Header.Get(t).Set(`Content-Type`, `application/json`)
	})
}

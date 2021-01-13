package httpspec

import (
	"github.com/adamluzsi/testcase"
)

func ContentTypeIsJSON(s *testcase.Spec) {
	s.Before(func(t *testcase.T) {
		HeaderGet(t).Set(`Content-Type`, `application/json`)
	})
}

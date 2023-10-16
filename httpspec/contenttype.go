package httpspec

import (
	"go.llib.dev/testcase"
)

func ContentTypeIsJSON(s *testcase.Spec) {
	s.Before(func(t *testcase.T) {
		Header.Get(t).Set(`Content-Type`, `application/json`)
	})
}

package httpspec

import (
	"net/http"
	"net/http/httptest"

	"github.com/adamluzsi/testcase"
)

func LetResponseRecorder(s *testcase.Spec) testcase.Var[*httptest.ResponseRecorder] {
	return testcase.Let[*httptest.ResponseRecorder](s, func(t *testcase.T) *httptest.ResponseRecorder {
		return httptest.NewRecorder()
	})
}

func LetInboundRequest(s *testcase.Spec) testcase.Var[*http.Request] {
	return testcase.Let(s, func(t *testcase.T) *http.Request {
		return httptest.NewRequest(http.MethodGet, "/", nil)
	})
}

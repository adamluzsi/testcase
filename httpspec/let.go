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

func LetServer(s *testcase.Spec, handler testcase.VarInitFunc[http.Handler]) testcase.Var[*httptest.Server] {
	return testcase.Let(s, func(t *testcase.T) *httptest.Server {
		srv := httptest.NewServer(handler(t))
		t.Defer(srv.Close)
		return srv
	})
}

package httpspec

import (
	"net/http"
	"net/http/httptest"
	"net/url"

	"github.com/adamluzsi/testcase"
)

func LetServer(s *testcase.Spec, handler testcase.VarInit[http.Handler]) testcase.Var[*httptest.Server] {
	return testcase.Let(s, func(t *testcase.T) *httptest.Server {
		srv := httptest.NewServer(handler(t))
		t.Defer(srv.Close)
		return srv
	})
}

func ClientDo(t *testcase.T, srv *httptest.Server, r *http.Request) (*http.Response, error) {
	r = r.Clone(r.Context())
	us, err := url.Parse(srv.URL)
	t.Must.NoError(err)
	r.URL.Scheme = us.Scheme
	r.URL.Host = us.Host
	r.RequestURI = ""
	return srv.Client().Do(r)
}

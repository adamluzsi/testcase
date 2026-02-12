package tchttp

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"

	"go.llib.dev/testcase"
	"go.llib.dev/testcase/assert"
	"go.llib.dev/testcase/random"
)

func LetResponseRecorder(s *testcase.Spec) testcase.Var[*httptest.ResponseRecorder] {
	return testcase.Let(s, func(t *testcase.T) *httptest.ResponseRecorder {
		return httptest.NewRecorder()
	})
}

func LetClientRequest(s *testcase.Spec, rvs ...RequestOption) testcase.Var[*http.Request] {
	s.H().Helper()
	rv := mergeRO(s, rvs...)
	return testcase.Let(s, func(t *testcase.T) *http.Request {
		defer func() { assert.Must(t).Nil(recover()) }()
		r, err := http.NewRequestWithContext(rv.Context.Get(t), rv.Method.Get(t), rv.url(t).String(), rv.body(t))
		assert.Must(t).NoError(err)
		r.Header = rv.Header.Get(t)
		return r
	})
}

func LetServerRequest(s *testcase.Spec, rvs ...RequestOption) testcase.Var[*http.Request] {
	s.H().Helper()
	rv := mergeRO(s, rvs...)
	return testcase.Let(s, func(t *testcase.T) *http.Request {
		defer func() { assert.Must(t).Nil(recover()) }() // catch httptest.NewRequest panic and fail the test
		r := httptest.NewRequest(rv.Method.Get(t), rv.url(t).String(), rv.body(t))
		r = r.WithContext(rv.Context.Get(t))
		r.Header = rv.Header.Get(t)
		r.Host = rv.Host.Get(t)
		return r
	})
}

type RequestOption struct {
	Context testcase.Var[context.Context]
	Scheme  testcase.Var[string]
	Host    testcase.Var[string]
	Method  testcase.Var[string]
	Path    testcase.Var[string]
	Query   testcase.Var[url.Values]
	Header  testcase.Var[http.Header]
	Body    testcase.Var[any]
}

func mergeRO(s *testcase.Spec, rvs ...RequestOption) RequestOption {
	var ro RequestOption
	for _, rv := range rvs {
		ro.merge(rv)
	}
	ro.init(s)
	return ro
}

func (o *RequestOption) merge(oth RequestOption) {
	o.Context = cmpVarOr(oth.Context, o.Context)
	o.Scheme = cmpVarOr(oth.Scheme, o.Scheme)
	o.Host = cmpVarOr(oth.Host, o.Host)
	o.Method = cmpVarOr(oth.Method, o.Method)
	o.Path = cmpVarOr(oth.Path, o.Path)
	o.Query = cmpVarOr(oth.Query, o.Query)
	o.Header = cmpVarOr(oth.Header, o.Header)
	o.Body = cmpVarOr(oth.Body, o.Body)
}

func cmpVarInitOr[T any](v, oth testcase.VarInit[T]) testcase.VarInit[T] {
	if v != nil {
		return v
	}
	return oth
}

func cmpVarOr[T any](v, oth testcase.Var[T]) testcase.Var[T] {
	if v.ID != "" {
		return v
	}
	return oth
}

func (o RequestOption) body(t *testcase.T) io.Reader {
	return asIOReader(t, o.Header.Get(t), o.Body.Get(t))
}

func (o RequestOption) url(t *testcase.T) *url.URL {
	return &url.URL{
		Scheme:   o.Scheme.Get(t),
		Host:     o.Host.Get(t),
		Path:     o.Path.Get(t),
		RawPath:  o.Path.Get(t),
		RawQuery: o.Query.Get(t).Encode(),
	}
}

func (o *RequestOption) init(s *testcase.Spec) {
	s.H().Helper()
	if o.Context.ID == "" {
		o.Context = testcase.Let(s, func(t *testcase.T) context.Context {
			return context.Background()
		})
	}
	if o.Scheme.ID == "" {
		o.Scheme = testcase.Let(s, func(t *testcase.T) string {
			return "http"
		})
	}
	if o.Host.ID == "" {
		o.Host = testcase.Let(s, func(t *testcase.T) string {
			return fmt.Sprintf("www.%s.com", t.Random.StringNC(5, random.CharsetAlpha()))
		})
	}
	if o.Method.ID == "" {
		o.Method = testcase.LetValue(s, http.MethodGet)
	}
	if o.Path.ID == "" {
		o.Path = testcase.LetValue(s, "/")
	}
	if o.Query.ID == "" {
		o.Query = testcase.Let(s, func(t *testcase.T) url.Values {
			return make(url.Values)
		})
	}
	if o.Header.ID == "" {
		o.Header = testcase.Let(s, func(t *testcase.T) http.Header {
			return make(http.Header)
		})
	}
	if o.Body.ID == "" {
		o.Body = testcase.LetValue[any](s, nil)
	}
}

func LetServer(s *testcase.Spec, handler testcase.VarInit[http.Handler]) testcase.Var[*httptest.Server] {
	return testcase.Let(s, func(t *testcase.T) *httptest.Server {
		srv := httptest.NewServer(handler(t))
		t.Defer(srv.Close)
		return srv
	})
}

func ServerClientDo(t *testcase.T, srv *httptest.Server, r *http.Request) (*http.Response, error) {
	r = r.Clone(r.Context())
	us, err := url.Parse(srv.URL)
	assert.Must(t).NoError(err)
	r.URL.Scheme = us.Scheme
	r.URL.Host = us.Host
	r.RequestURI = ""
	return srv.Client().Do(r)
}

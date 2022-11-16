package httpspec

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"

	"github.com/adamluzsi/testcase"
	"github.com/adamluzsi/testcase/random"
)

func LetResponseRecorder(s *testcase.Spec) testcase.Var[*httptest.ResponseRecorder] {
	return testcase.Let[*httptest.ResponseRecorder](s, func(t *testcase.T) *httptest.ResponseRecorder {
		return httptest.NewRecorder()
	})
}

func LetRequest(s *testcase.Spec, rv RequestVar) testcase.Var[*http.Request] {
	rv = rv.withDefaults(s)
	return testcase.Let(s, func(t *testcase.T) *http.Request {
		defer func() { t.Must.Nil(recover()) }()
		u := url.URL{
			Scheme:   rv.Scheme.Get(t),
			Host:     rv.Host.Get(t),
			Path:     rv.Path.Get(t),
			RawPath:  rv.Path.Get(t),
			RawQuery: rv.Query.Get(t).Encode(),
		}
		r, err := http.NewRequestWithContext(rv.Context.Get(t), rv.Method.Get(t), u.String(), asIOReader(t, rv.Header.Get(t), rv.Body.Get(t)))
		t.Must.NoError(err)
		r.Header = rv.Header.Get(t)
		return r
	})
}

type RequestVar struct {
	Context testcase.Var[context.Context]
	Scheme  testcase.Var[string]
	Host    testcase.Var[string]
	Method  testcase.Var[string]
	Path    testcase.Var[string]
	Query   testcase.Var[url.Values]
	Header  testcase.Var[http.Header]
	Body    testcase.Var[any]
}

func (rv RequestVar) withDefaults(s *testcase.Spec) RequestVar {
	if rv.Context.ID == "" {
		rv.Context = testcase.Let(s, func(t *testcase.T) context.Context {
			return context.Background()
		})
	}
	if rv.Scheme.ID == "" {
		rv.Scheme = testcase.Let(s, func(t *testcase.T) string {
			return t.Random.SliceElement([]string{"http", "https"}).(string)
		})
	}
	if rv.Host.ID == "" {
		rv.Host = testcase.Let(s, func(t *testcase.T) string {
			return fmt.Sprintf("www.%s.com", t.Random.StringNC(5, random.CharsetAlpha()))
		})
	}
	if rv.Method.ID == "" {
		rv.Method = testcase.LetValue(s, http.MethodGet)
	}
	if rv.Path.ID == "" {
		rv.Path = testcase.LetValue(s, "/")
	}
	if rv.Query.ID == "" {
		rv.Query = testcase.Let(s, func(t *testcase.T) url.Values {
			return make(url.Values)
		})
	}
	if rv.Header.ID == "" {
		rv.Header = testcase.Let(s, func(t *testcase.T) http.Header {
			return make(http.Header)
		})
	}
	if rv.Body.ID == "" {
		rv.Body = testcase.LetValue[any](s, nil)
	}
	return rv
}

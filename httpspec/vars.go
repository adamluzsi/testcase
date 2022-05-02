package httpspec

import (
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"

	"github.com/adamluzsi/testcase"
)

var (
	// Handler prepares the current testcase spec scope to be ready for http handler testing.
	// You define your spec subject with this and all the request will be pointed towards this.
	Handler = testcase.Var[http.Handler]{ID: `httpspec:Handler`}
	// Context allow to retrieve the current test scope's request context.
	Context = testcase.Var[context.Context]{ID: `httpspec:Context`, Init: func(t *testcase.T) context.Context {
		return context.Background()
	}}
	Method = testcase.Var[string]{ID: `httpspec:Method`, Init: func(t *testcase.T) string {
		return http.MethodGet
	}}
	Path = testcase.Var[string]{ID: `httpspec:Path`, Init: func(t *testcase.T) string {
		return `/`
	}}
	// Query allows you to retrieve the current test scope's http PathGet query that will be used for ServeHTTP.
	// In a Before Block you can access the query and then specify the values in it.
	Query = testcase.Var[url.Values]{ID: `httpspec:QueryGet`, Init: func(t *testcase.T) url.Values {
		return url.Values{}
	}}
	// Header allows you to set the current test scope's http PathGet for ServeHTTP.
	Header = testcase.Var[http.Header]{ID: `httpspec:Header.Get`, Init: func(t *testcase.T) http.Header {
		return http.Header{}
	}}
)

var (
	Request = testcase.Var[*http.Request]{
		ID: "httpspec:Request",
		Init: func(t *testcase.T) *http.Request {
			target, _ := url.Parse(Path.Get(t))
			target.RawQuery = Query.Get(t).Encode()
			if isDebugEnabled(t) {
				t.Log(`http.Request.Method:`, Method.Get(t))
				t.Log(`http.Request.Path`, target.String())
			}
			r := httptest.NewRequest(Method.Get(t), target.String(), bodyAsIOReader(t))
			r = r.WithContext(Context.Get(t))
			r.Header = Header.Get(t)
			return r
		},
	}
	ResponseRecorder = testcase.Var[*httptest.ResponseRecorder]{
		ID: "httpspec:ResponseRecorder",
		Init: func(t *testcase.T) *httptest.ResponseRecorder {
			return httptest.NewRecorder()
		},
	}
)

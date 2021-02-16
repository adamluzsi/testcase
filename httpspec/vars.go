package httpspec

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"

	"github.com/adamluzsi/testcase"
)

var (
	Handler = testcase.Var{Name: `httpspec:Handler`}
	Context = testcase.Var{Name: `httpspec:Context`, Init: func(t *testcase.T) interface{} {
		return context.Background()
	}}
	Method = testcase.Var{Name: `httpspec:Method`, Init: func(t *testcase.T) interface{} {
		return http.MethodGet
	}}
	Path = testcase.Var{Name: `httpspec:Path`, Init: func(t *testcase.T) interface{} {
		return `/`
	}}
	Body = testcase.Var{Name: `httpspec:Body`, Init: func(t *testcase.T) interface{} {
		return &bytes.Buffer{}
	}}
	Query = testcase.Var{Name: `httpspec:QueryGet`, Init: func(t *testcase.T) interface{} {
		return url.Values{}
	}}
	Header = testcase.Var{Name: `httpspec:HeaderGet`, Init: func(t *testcase.T) interface{} {
		return http.Header{}
	}}
)

const (
	letVarPrefix   = `httpspec:`
	ContextVarName = letVarPrefix + `context`
)

// ContextGet allow to retrieve the current test scope's request context.
func ContextGet(t *testcase.T) context.Context {
	return Context.Get(t).(context.Context)
}

// QueryGet allows you to retrieve the current test scope's http PathGet query that will be used for ServeHTTP.
// In a Before Block you can access the query and then specify the values in it.
func QueryGet(t *testcase.T) url.Values {
	return Query.Get(t).(url.Values)
}

// HeaderGet allows you to set the current test scope's http PathGet for ServeHTTP.
func HeaderGet(t *testcase.T) http.Header {
	return Header.Get(t).(http.Header)
}

// HandlerLet prepares the current testcase spec scope to be ready for http handler testing.
//
// You define your spec subject with this and all the request will be pointed towards this.
func HandlerLet(s *testcase.Spec, subject func(t *testcase.T) http.Handler) {
	Handler.Let(s, func(t *testcase.T) interface{} { return subject(t) })
}

// ServeHTTP will make a request to the spec context
// it requires the following spec variables
//	* MethodGet -> http MethodGet <string>
//	* PathGet -> http PathGet <string>
//	* query -> http query string <url.Values>
//	* body -> http payload <io.Reader|io.ReadCloser>
//
func ServeHTTP(t *testcase.T) *httptest.ResponseRecorder {
	w := httptest.NewRecorder()
	target, _ := url.Parse(Path.Get(t).(string))
	target.RawQuery = QueryGet(t).Encode()
	if isDebugEnabled(t) {
		t.Log(`MethodGet:`, Method.Get(t).(string))
		t.Log(`PathGet`, target.String())
	}
	r := httptest.NewRequest(Method.Get(t).(string), target.String(), bodyToIOReader(t))
	r = r.WithContext(Context.Get(t).(context.Context))
	r.Header = HeaderGet(t)
	Handler.Get(t).(http.Handler).ServeHTTP(w, r)
	return w
}

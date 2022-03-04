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
	Body = testcase.Var[any]{ID: `httpspec:Body`, Init: func(t *testcase.T) any {
		return &bytes.Buffer{}
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

// ServeHTTP will make a request to the spec context
// it requires the following spec variables
//	* MethodGet -> http MethodGet <string>
//	* PathGet -> http PathGet <string>
//	* query -> http query string <url.Values>
//	* body -> http payload <io.Reader|io.ReadCloser>
//
func ServeHTTP(t *testcase.T) *httptest.ResponseRecorder {
	w := httptest.NewRecorder()
	target, _ := url.Parse(Path.Get(t))
	target.RawQuery = Query.Get(t).Encode()
	if isDebugEnabled(t) {
		t.Log(`MethodGet:`, Method.Get(t))
		t.Log(`PathGet`, target.String())
	}
	r := httptest.NewRequest(Method.Get(t), target.String(), bodyToIOReader(t))
	r = r.WithContext(Context.Get(t))
	r.Header = Header.Get(t)
	Handler.Get(t).ServeHTTP(w, r)
	return w
}

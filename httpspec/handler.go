package httpspec

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"

	"github.com/adamluzsi/testcase"
)

// ServeHTTP will make a request to the spec context
// it requires the following spec variables
//	* method -> http method <string>
//	* path -> http path <string>
//	* query -> http query string <url.Values>
//	* body -> http payload <io.Reader|io.ReadCloser>
//
func ServeHTTP(t *testcase.T) *httptest.ResponseRecorder {
	w := httptest.NewRecorder()
	target, _ := url.Parse(path(t))
	target.RawQuery = Query(t).Encode()
	if isDebugEnabled(t) {
		t.Log(`method:`, method(t))
		t.Log(`path`, target.String())
	}
	r := httptest.NewRequest(method(t), target.String(), bodyToIOReader(t))
	r = r.WithContext(ctx(t))
	r.Header = Header(t)
	handler(t).ServeHTTP(w, r)
	return w
}

// HandlerSpec prepares the current testcase spec scope to be ready for http handler testing.
//
// You define your spec subject with this and all the request will be pointed towards this.
func HandlerSpec(s *testcase.Spec, subject func(t *testcase.T) http.Handler) {
	setupDebug(s)
	letHandler(s, subject)
	letHeader(s, func(t *testcase.T) http.Header { return http.Header{} })
	letQuery(s, func(t *testcase.T) url.Values { return url.Values{} })
	LetContext(s, func(t *testcase.T) context.Context { return context.Background() })
	LetMethod(s, func(t *testcase.T) string { return http.MethodGet })
	LetPath(s, func(t *testcase.T) string { return `/` })
	LetBody(s, func(t *testcase.T) interface{} { return &bytes.Buffer{} })
}

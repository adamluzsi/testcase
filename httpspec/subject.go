package httpspec

import (
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"

	"github.com/adamluzsi/testcase"
)

var Subject = testcase.Var{
	Name: `httpspec:SubjectGet`,
	Init: func(t *testcase.T) interface{} { return Act(t) },
}

func SubjectGet(t *testcase.T) *httptest.ResponseRecorder {
	return Subject.Get(t).(*httptest.ResponseRecorder)
}

// SubjectLet prepares the current testcase spec scope to be ready for http handler testing.
//
// You define your spec subject with this and all the request will be pointed towards this.
func SubjectLet(s *testcase.Spec, subject func(t *testcase.T) http.Handler) {
	Handler.Let(s, func(t *testcase.T) interface{} { return subject(t) })
}

// Act will make a request to the spec context
// it requires the following spec variables
//	* MethodGet -> http MethodGet <string>
//	* PathGet -> http PathGet <string>
//	* query -> http query string <url.Values>
//	* body -> http payload <io.Reader|io.ReadCloser>
//
func Act(t *testcase.T) *httptest.ResponseRecorder {
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
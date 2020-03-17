package httpspec_test

import (
	"net/http"

	"github.com/adamluzsi/testcase"
	"github.com/adamluzsi/testcase/httpspec"
)

func ExampleLetBody() {
	s := testcase.NewSpec(testingT)

	httpspec.GivenThisIsAnAPI(s)
	httpspec.LetHandler(s, func(t *testcase.T) http.Handler { return MyHandler{} })

	httpspec.LetBody(s, func(t *testcase.T) interface{} {
		return map[string]string{"hello": "world"}
	})

	s.Before(func(t *testcase.T) {
		// this set the content-type for json, so json marshal will be used.
		httpspec.Header(t).Set(`Content-Type`, `application/json`)
	})

	s.Test(`the *http.Request body io.Reader will have the encoded body`, func(t *testcase.T) {
		httpspec.ServeHTTP(t)
	})
}

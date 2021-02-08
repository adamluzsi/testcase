package httpspec_test

import (
	"net/http"

	"github.com/adamluzsi/testcase"
	"github.com/adamluzsi/testcase/httpspec"
)

func ExampleQuery() {
	s := testcase.NewSpec(testingT)

	httpspec.HandlerLet(s, func(t *testcase.T) http.Handler { return MyHandler{} })

	s.Before(func(t *testcase.T) {
		// this is ideal to represent query string inputs
		httpspec.QueryGet(t).Set(`foo`, `bar`)
	})

	s.Test(`the *http.Request URL QueryGet will have foo=bar`, func(t *testcase.T) {
		httpspec.ServeHTTP(t)
	})
}

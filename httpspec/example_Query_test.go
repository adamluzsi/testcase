package httpspec_test

import (
	"net/http"

	"go.llib.dev/testcase"
	"go.llib.dev/testcase/httpspec"
)

func ExampleQuery() {
	s := testcase.NewSpec(testingT)

	httpspec.Handler.Let(s, func(t *testcase.T) http.Handler { return MyHandler{} })

	s.Before(func(t *testcase.T) {
		// this is ideal to represent query string inputs
		httpspec.Query.Get(t).Set(`foo`, `bar`)
	})

	s.Test(`the *http.InboundRequest URL QueryGet will have foo=bar`, func(t *testcase.T) {
		httpspec.ServeHTTP(t)
	})
}

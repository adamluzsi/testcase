package httpspec_test

import (
	"net/http"

	"go.llib.dev/testcase"
	"go.llib.dev/testcase/httpspec"
)

func ExampleLetMethodValue() {
	s := testcase.NewSpec(testingT)

	httpspec.Handler.Let(s, func(t *testcase.T) http.Handler {
		return MyHandler{}
	})

	// set the HTTP Method to get for the *http.InboundRequest
	httpspec.Method.LetValue(s, http.MethodGet)

	s.Test(`GET /`, func(t *testcase.T) {
		httpspec.ServeHTTP(t)
	})
}

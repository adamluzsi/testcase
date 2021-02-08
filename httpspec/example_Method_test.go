package httpspec_test

import (
	"net/http"

	"github.com/adamluzsi/testcase"
	"github.com/adamluzsi/testcase/httpspec"
)

func ExampleLetMethodValue() {
	s := testcase.NewSpec(testingT)

	httpspec.HandlerLet(s, func(t *testcase.T) http.Handler {
		return MyHandler{}
	})

	// set the HTTP Method to get for the *http.Request
	httpspec.Method.LetValue(s, http.MethodGet)

	s.Test(`GET /`, func(t *testcase.T) {
		httpspec.ServeHTTP(t)
	})
}

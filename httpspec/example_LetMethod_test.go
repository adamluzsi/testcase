package httpspec_test

import (
	"net/http"

	"github.com/adamluzsi/testcase"
	"github.com/adamluzsi/testcase/httpspec"
)

func ExampleLetMethod() {
	s := testcase.NewSpec(testingT)

	httpspec.GivenThisIsAnAPI(s)

	httpspec.LetMethod(s, func(t *testcase.T) string {
		// set the HTTP Method to get for the *http.Request
		return http.MethodGet
	})

	s.Test(`GET /`, func(t *testcase.T) {
		httpspec.ServeHTTP(t)
	})
}

func ExampleLetMethodValue() {
	s := testcase.NewSpec(testingT)

	httpspec.GivenThisIsAnAPI(s)

	// set the HTTP Method to get for the *http.Request
	httpspec.LetMethodValue(s, http.MethodGet)

	s.Test(`GET /`, func(t *testcase.T) {
		httpspec.ServeHTTP(t)
	})
}

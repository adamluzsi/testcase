package httpspec_test

import (
	"net/http"

	"github.com/adamluzsi/testcase"
	"github.com/adamluzsi/testcase/httpspec"
)

func ExampleHeader() {
	s := testcase.NewSpec(testingT)

	httpspec.HandlerSpec(s, func(t *testcase.T) http.Handler { return MyHandler{} })

	s.Before(func(t *testcase.T) {
		// this is ideal to represent query string inputs
		httpspec.Header(t).Set(`Foo`, `bar`)
	})

	s.Test(`the *http.Request URL Query will have 'Foo: bar'`, func(t *testcase.T) {
		httpspec.ServeHTTP(t)
	})
}

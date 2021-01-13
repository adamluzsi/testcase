package httpspec_test

import (
	"net/http"

	"github.com/adamluzsi/testcase"
	"github.com/adamluzsi/testcase/httpspec"
)

func ExampleHeader() {
	s := testcase.NewSpec(testingT)

	httpspec.SubjectLet(s, func(t *testcase.T) http.Handler { return MyHandler{} })

	s.Before(func(t *testcase.T) {
		// this is ideal to represent query string inputs
		httpspec.HeaderGet(t).Set(`Foo`, `bar`)
	})

	s.Test(`the *http.Request URL QueryGet will have 'Foo: bar'`, func(t *testcase.T) {
		httpspec.SubjectGet(t)
	})
}

package httpspec_test

import (
	"context"
	"net/http"

	"github.com/adamluzsi/testcase"
	"github.com/adamluzsi/testcase/httpspec"
)

func ExampleLetContext() {
	s := testcase.NewSpec(testingT)

	httpspec.GivenThisIsAnAPI(s)
	httpspec.LetHandler(s, func(t *testcase.T) http.Handler { return MyHandler{} })

	httpspec.LetContext(s, func(t *testcase.T) context.Context {
		return context.Background()
	})

	s.Test(`the *http.Request#Context() will have foo-bar`, func(t *testcase.T) {
		httpspec.ServeHTTP(t)
	})
}

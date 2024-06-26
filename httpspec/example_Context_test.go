package httpspec_test

import (
	"context"
	"net/http"

	"go.llib.dev/testcase"
	"go.llib.dev/testcase/httpspec"
)

func ExampleContext_withValue() {
	s := testcase.NewSpec(testingT)

	httpspec.Handler.Let(s, func(t *testcase.T) http.Handler { return MyHandler{} })

	s.Before(func(t *testcase.T) {
		// This approach can help you representing middleware prerequisites.
		// Use httpspec.Context.Set only if you can't solve your goal
		// with httpspec.Context.Let or httpspec.Context.LetValue.
		httpspec.Context.Set(t, context.WithValue(httpspec.Context.Get(t), `foo`, `bar`))
	})

	s.Test(`the *http.InboundRequest#Context() will have foo-bar`, func(t *testcase.T) {
		httpspec.ServeHTTP(t)
	})
}

package httpspec_test

import (
	"context"
	"net/http"

	"github.com/adamluzsi/testcase"
	"github.com/adamluzsi/testcase/httpspec"
)

func ExampleLetContext_withValue() {
	s := testcase.NewSpec(testingT)

	httpspec.GivenThisIsAnAPI(s)
	httpspec.LetHandler(s, func(t *testcase.T) http.Handler { return MyHandler{} })

	s.Before(func(t *testcase.T) {
		// this is ideal for representing middleware prerequisite
		// in the form of a value in the context that is guaranteed by a middleware.
		// Use this only if you cannot make it part of the specification level context value deceleration with LetContext.
		ctx := t.I(httpspec.RequestContextVarName).(context.Context)
		ctx = context.WithValue(ctx, `foo`, `bar`)
		t.Let(httpspec.RequestContextVarName, ctx)
	})

	s.Test(`the *http.Request#Context() will have foo-bar`, func(t *testcase.T) {
		httpspec.ServeHTTP(t)
	})
}

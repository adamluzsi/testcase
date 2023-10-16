package httpspec_test

import (
	"net/http"
	"testing"

	"go.llib.dev/testcase"
	"go.llib.dev/testcase/httpspec"
)

func TestItBehavesLikeHandlerMiddleware(t *testing.T) {
	s := testcase.NewSpec(t)
	httpspec.ItBehavesLikeHandlerMiddleware(s, func(t *testcase.T, next http.Handler) http.Handler {
		return ExampleHandler{Next: next}
	})
}

func TestHandlerMiddlewareContract_Spec(t *testing.T) {
	testcase.RunSuite(t, httpspec.HandlerMiddlewareContract{
		Subject: func(t *testcase.T, next http.Handler) http.Handler {
			return ExampleHandler{Next: next}
		},
	})
}

type ExampleHandler struct {
	Next http.Handler
}

func (h ExampleHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h.Next.ServeHTTP(w, r)
}

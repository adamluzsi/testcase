package httpspec_test

import (
	"net/http"
	"testing"

	"go.llib.dev/testcase"
	"go.llib.dev/testcase/spec/httpspec"
)

func TestHandlerMiddleware(t *testing.T) {
	s := testcase.NewSpec(t)

	httpspec.HandlerMiddleware(func(t *testcase.T, next http.Handler) http.Handler {
		return ExampleHandler{Next: next}
	}).Spec(s)
}

type ExampleHandler struct {
	Next http.Handler
}

func (h ExampleHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h.Next.ServeHTTP(w, r)
}

func ExampleRoundTripperMiddleware() {
	var tb testing.TB

	s := testcase.NewSpec(tb)

	contract := httpspec.RoundTripperMiddleware(func(t *testcase.T, next http.RoundTripper) http.RoundTripper {
		return ExampleRoundTripper{Next: next}
	})

	s.Context("it behaves as you would expect from an http RoundTripper middleware",
		contract.Spec)
}

func TestRoundTripperMiddleware(t *testing.T) {
	s := testcase.NewSpec(t)
	httpspec.RoundTripperMiddleware(func(t *testcase.T, next http.RoundTripper) http.RoundTripper {
		return ExampleRoundTripper{Next: next}
	}).Spec(s)
}

type ExampleRoundTripper struct {
	Next http.RoundTripper
}

func (rt ExampleRoundTripper) RoundTrip(r *http.Request) (*http.Response, error) {
	return rt.Next.RoundTrip(r)
}

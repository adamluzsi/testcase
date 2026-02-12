package tchttp_test

import (
	"net/http"
	"testing"

	"go.llib.dev/testcase"
	"go.llib.dev/testcase/assert"
	"go.llib.dev/testcase/let"
	"go.llib.dev/testcase/tchttp"
)

func TestHandlerMiddleware(t *testing.T) {
	s := testcase.NewSpec(t)

	tchttp.HandlerMiddleware(func(t *testcase.T, next http.Handler) http.Handler {
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

	contract := tchttp.RoundTripperMiddleware(func(t *testcase.T, next http.RoundTripper) http.RoundTripper {
		return ExampleRoundTripper{Next: next}
	})

	s.Context("it behaves as you would expect from an http RoundTripper middleware",
		contract.Spec)
}

func TestRoundTripperMiddleware(t *testing.T) {
	s := testcase.NewSpec(t)

	s.Context("smoke",
		tchttp.RoundTripperMiddleware(func(t *testcase.T, next http.RoundTripper) http.RoundTripper {
			return ExampleRoundTripper{Next: next}
		}).Spec)

	s.Context("when round tripper middleware got response injected as a config", func(s *testcase.Spec) {
		yielded := let.VarOf(s, false)

		response := let.Var(s, func(t *testcase.T) *http.Response {
			yielded.Set(t, true)
			return &http.Response{StatusCode: http.StatusOK}
		})

		var ok bool
		s.After(func(t *testcase.T) {
			if yielded.Get(t) {
				ok = true
			}
		})

		s.AfterAll(func(tb testing.TB) {
			assert.True(t, ok, "expecte that response option was used")
		})

		tchttp.RoundTripperMiddleware(func(t *testcase.T, next http.RoundTripper) http.RoundTripper {
			return ExampleRoundTripper{Next: next}
		}, tchttp.RoundTripperMiddlewareOption{Response: response.Get}).Spec(s)
	})
}

type ExampleRoundTripper struct {
	Next http.RoundTripper
}

func (rt ExampleRoundTripper) RoundTrip(r *http.Request) (*http.Response, error) {
	return rt.Next.RoundTrip(r)
}

package httpspec_test

import (
	"net/http"
	"testing"

	"github.com/adamluzsi/testcase"
	"github.com/adamluzsi/testcase/httpspec"
)

func TestRoundTripperContract_Spec(t *testing.T) {
	testcase.RunContract(t, httpspec.RoundTripperContract{
		Subject: func(t *testcase.T, next http.RoundTripper) http.RoundTripper {
			return ExampleRoundTripper{Next: next}
		},
	})
}

type ExampleRoundTripper struct {
	Next http.RoundTripper
}

func (rt ExampleRoundTripper) RoundTrip(r *http.Request) (*http.Response, error) {
	return rt.Next.RoundTrip(r)
}

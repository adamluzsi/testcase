package fihttp_test

import (
	"errors"
	"net"
	"net/http"
	"syscall"
	"testing"

	"github.com/adamluzsi/testcase"
	"github.com/adamluzsi/testcase/faultinject"
	"github.com/adamluzsi/testcase/faultinject/fihttp"
	"github.com/adamluzsi/testcase/httpspec"
	"github.com/adamluzsi/testcase/random"
)

func TestRoundTripper(t *testing.T) {
	s := testcase.NewSpec(t)

	next := testcase.Let(s, func(t *testcase.T) http.RoundTripper {
		return httpspec.StubRoundTripper(func(r *http.Request) (*http.Response, error) {
			resp := &http.Response{StatusCode: t.Random.ElementFromSlice([]int{
				http.StatusOK,
				http.StatusTeapot,
				http.StatusBadRequest,
				http.StatusInternalServerError,
			}).(int)}
			err := errors.New(t.Random.String())
			return resp, err
		})
	})
	serviceName := testcase.Let(s, func(t *testcase.T) string {
		return ""
	})
	newRoundTripper := func(t *testcase.T, next http.RoundTripper) http.RoundTripper {
		return &fihttp.RoundTripper{
			Next:        next,
			ServiceName: serviceName.Get(t),
		}
	}
	roundTripper := testcase.Let(s, func(t *testcase.T) *fihttp.RoundTripper {
		return newRoundTripper(t, next.Get(t)).(*fihttp.RoundTripper)
	})

	s.Describe(".RoundTrip", func(s *testcase.Spec) {
		request := httpspec.OutboundRequest.Bind(s)
		subject := func(t *testcase.T) (*http.Response, error) {
			return roundTripper.Get(t).RoundTrip(request.Get(t))
		}

		httpspec.ItBehavesLikeRoundTripperMiddleware(s, newRoundTripper)

		s.When("net-timeout error is injected", func(s *testcase.Spec) {
			s.Before(func(t *testcase.T) {
				r := request.Get(t)
				r = r.WithContext(faultinject.Inject(r.Context(), fihttp.TagTimeout{}))
				request.Set(t, r)
			})

			thenItWillTimeoutErr(s, newRoundTripper, subject)
		})

		s.When("net-timeout error is injected for a given service", func(s *testcase.Spec) {
			targetServiceName := testcase.Let(s, func(t *testcase.T) string {
				return t.Random.StringNC(4, random.CharsetAlpha())
			})

			s.Before(func(t *testcase.T) {
				r := request.Get(t)
				r = r.WithContext(faultinject.Inject(r.Context(), fihttp.TagTimeout{ServiceName: targetServiceName.Get(t)}))
				request.Set(t, r)
			})

			s.And("the round tripper belongs meant for that service", func(s *testcase.Spec) {
				serviceName.Let(s, func(t *testcase.T) string {
					return targetServiceName.Get(t)
				})

				thenItWillTimeoutErr(s, newRoundTripper, subject)
			})

			s.And("the our round tripper meant for a different service", func(s *testcase.Spec) {
				serviceName.Let(s, func(t *testcase.T) string {
					return t.Random.StringNC(5, random.CharsetAlpha())
				})

				httpspec.ItBehavesLikeRoundTripperMiddleware(s, newRoundTripper)
			})

			s.And("the our round tripper doesn't have a service name specified to it", func(s *testcase.Spec) {
				serviceName.Let(s, func(t *testcase.T) string {
					return ""
				})

				httpspec.ItBehavesLikeRoundTripperMiddleware(s, newRoundTripper)
			})
		})

		s.When("connection-refused error is injected", func(s *testcase.Spec) {
			s.Before(func(t *testcase.T) {
				r := request.Get(t)
				r = r.WithContext(faultinject.Inject(r.Context(), fihttp.TagConnectionRefused{}))
				request.Set(t, r)
			})

			thenItWillConnectionRefuseErr(s, newRoundTripper, subject)
		})

		s.When("connection-refused error is injected for a given service", func(s *testcase.Spec) {
			targetServiceName := testcase.Let(s, func(t *testcase.T) string {
				return t.Random.StringNC(4, random.CharsetAlpha())
			})

			s.Before(func(t *testcase.T) {
				r := request.Get(t)
				r = r.WithContext(faultinject.Inject(r.Context(), fihttp.TagConnectionRefused{ServiceName: targetServiceName.Get(t)}))
				request.Set(t, r)
			})

			s.And("the round tripper belongs meant for that service", func(s *testcase.Spec) {
				serviceName.Let(s, func(t *testcase.T) string {
					return targetServiceName.Get(t)
				})

				thenItWillConnectionRefuseErr(s, newRoundTripper, subject)
			})

			s.And("the our round tripper meant for a different service", func(s *testcase.Spec) {
				serviceName.Let(s, func(t *testcase.T) string {
					return t.Random.StringNC(5, random.CharsetAlpha())
				})

				httpspec.ItBehavesLikeRoundTripperMiddleware(s, newRoundTripper)
			})

			s.And("the our round tripper doesn't have a service name specified to it", func(s *testcase.Spec) {
				serviceName.Let(s, func(t *testcase.T) string {
					return ""
				})

				httpspec.ItBehavesLikeRoundTripperMiddleware(s, newRoundTripper)
			})
		})
	})

}

func thenItWillTimeoutErr(s *testcase.Spec,
	newRoundTripper httpspec.MakeRoundTripperFunc,
	act func(t *testcase.T) (*http.Response, error),
) {
	s.Then("it will trigger a connection refused error", func(t *testcase.T) {
		response, err := act(t)
		t.Must.NotNil(err)
		t.Must.Nil(response)

		netErr, ok := err.(net.Error)
		t.Must.True(ok)
		t.Must.True(netErr.Timeout())
	})

	s.And("after exhausting the injected fault", func(s *testcase.Spec) {
		s.Before(func(t *testcase.T) {
			_, _ = act(t)
		})

		httpspec.ItBehavesLikeRoundTripperMiddleware(s, newRoundTripper)
	})
}

func thenItWillConnectionRefuseErr(s *testcase.Spec,
	newRoundTripper httpspec.MakeRoundTripperFunc,
	act func(t *testcase.T) (*http.Response, error),
) {
	s.Then("it will trigger a connection refused error", func(t *testcase.T) {
		response, err := act(t)
		t.Must.ErrorIs(syscall.ECONNREFUSED, err)
		t.Must.Nil(response)
	})

	s.And("after exhausting the injected fault", func(s *testcase.Spec) {
		s.Before(func(t *testcase.T) {
			_, _ = act(t)
		})

		httpspec.ItBehavesLikeRoundTripperMiddleware(s, newRoundTripper)
	})
}

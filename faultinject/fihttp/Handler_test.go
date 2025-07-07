package fihttp_test

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"

	"go.llib.dev/testcase/let"

	"go.llib.dev/testcase"
	"go.llib.dev/testcase/faultinject"
	"go.llib.dev/testcase/faultinject/fihttp"
	"go.llib.dev/testcase/httpspec"
	"go.llib.dev/testcase/random"
)

func TestHandler(t *testing.T) {
	s := testcase.NewSpec(t)

	s.Before(func(t *testcase.T) {
		faultinject.EnableForTest(t)
	})

	type faultKey struct{}

	expectedErrOnFaultKey := let.Error(s)

	lastRequest := testcase.Let[*http.Request](s, nil)
	next := testcase.Let(s, func(t *testcase.T) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			lastRequest.Set(t, r)
			w.WriteHeader(http.StatusTeapot)
		})
	})
	serviceName := let.StringNC(s, 5, random.CharsetASCII())
	faultName := let.StringNC(s, 5, random.CharsetASCII())
	newHandler := func(t *testcase.T, next http.Handler) http.Handler {
		return &fihttp.Handler{
			Next:        next,
			ServiceName: serviceName.Get(t),
			FaultsMapping: fihttp.FaultsMapping{
				faultName.Get(t): func(ctx context.Context) context.Context {
					return faultinject.Inject(ctx, faultKey{}, expectedErrOnFaultKey.Get(t))
				},
			},
		}
	}
	handler := testcase.Let(s, func(t *testcase.T) *fihttp.Handler {
		return newHandler(t, next.Get(t)).(*fihttp.Handler)
	})

	s.Describe(".ServeHTTP", func(s *testcase.Spec) {
		writer := httpspec.ResponseRecorder.Bind(s)
		request := httpspec.InboundRequest.Bind(s)
		act := func(t *testcase.T) {
			handler.Get(t).ServeHTTP(writer.Get(t), request.Get(t))
		}

		httpspec.ItBehavesLikeHandlerMiddleware(s, newHandler)

		s.When("fault injection header is used to inject error", func(s *testcase.Spec) {
			injectedFaultInHeader := testcase.Let[any](s, nil)

			s.Before(func(t *testcase.T) {
				data, err := json.Marshal(injectedFaultInHeader.Get(t))
				t.Must.Nil(err)
				httpspec.Header.Get(t).Set(fihttp.Header, string(data))
			})

			s.And("the header contains fault which meant to our service", func(s *testcase.Spec) {
				name := testcase.Let[string](s, nil)
				injectedFaultInHeader.Let(s, func(t *testcase.T) any {
					return fihttp.Fault{
						ServiceName: serviceName.Get(t),
						Name:        name.Get(t),
					}
				})

				s.And("the injected fault name is registered in the mapping", func(s *testcase.Spec) {
					name.Let(s, func(t *testcase.T) string {
						return faultName.Get(t)
					})

					s.Then("it will inject the fault into the request context", func(t *testcase.T) {
						act(t)

						t.Must.NotNil(lastRequest.Get(t))
						t.Must.Equal(expectedErrOnFaultKey.Get(t), lastRequest.Get(t).Context().Value(faultKey{}))
					})
				})

				s.And("the injected fault name is not registered in the mapping", func(s *testcase.Spec) {
					name.Let(s, func(t *testcase.T) string {
						return faultName.Get(t) + t.Random.StringNC(5, random.CharsetAlpha())
					})

					s.Then("it will ignore the injected fault", func(t *testcase.T) {
						act(t)

						t.Must.NotNil(lastRequest.Get(t))
						t.Must.Nil(lastRequest.Get(t).Context().Value(faultKey{}))
					})
				})
			})

			s.And("the header contains faults which doesn't meant to our service", func(s *testcase.Spec) {
				othServiceName := testcase.Let(s, func(t *testcase.T) string {
					return t.Random.StringNC(5, random.CharsetAlpha())
				})
				injectedFaultInHeader.Let(s, func(t *testcase.T) any {
					return fihttp.Fault{
						ServiceName: othServiceName.Get(t),
						Name:        faultName.Get(t),
					}
				})

				next.Let(s, func(t *testcase.T) http.Handler {
					return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

						obreq, err := http.NewRequestWithContext(r.Context(), http.MethodGet, "http://example.com/", nil)
						t.Must.Nil(err)

						_, _ = fihttp.RoundTripper{
							Next: httpspec.RoundTripperFunc(func(r *http.Request) (*http.Response, error) {
								lastRequest.Set(t, r)
								return httpspec.Response.Get(t), nil
							}),
						}.RoundTrip(obreq)

						w.WriteHeader(http.StatusTeapot)
					})
				})

				s.Then("it will propagate the fault injection for the RoundTripper's outbound request header", func(t *testcase.T) {
					act(t)

					t.Must.NotNil(lastRequest.Get(t))
					t.Must.Equal(1, len(lastRequest.Get(t).Header.Values(fihttp.Header)))

					t.Logf("%#v", lastRequest.Get(t).Header.Values(fihttp.Header))

					var faults []fihttp.Fault
					t.Must.Nil(json.Unmarshal([]byte(lastRequest.Get(t).Header.Get(fihttp.Header)), &faults))
					t.Must.Contains(faults, injectedFaultInHeader.Get(t))
				})
			})
		})
	})
}

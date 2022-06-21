package fihttp_test

import (
	"encoding/json"
	"errors"
	"net/http"
	"testing"

	"github.com/adamluzsi/testcase"
	"github.com/adamluzsi/testcase/faultinject"
	"github.com/adamluzsi/testcase/faultinject/fihttp"
	"github.com/adamluzsi/testcase/httpspec"
	"github.com/adamluzsi/testcase/random"
)

func TestHandler(t *testing.T) {
	s := testcase.NewSpec(t)

	s.Before(func(t *testcase.T) {
		faultinject.EnableForTest(t)
	})

	type faultKey struct{}

	expectedErrOnFaultKey := testcase.Let(s, func(t *testcase.T) error {
		return errors.New(t.Random.String())
	})
	injector := testcase.Let(s, func(t *testcase.T) faultinject.Injector {
		return faultinject.Injector{}.OnTag(faultKey{}, expectedErrOnFaultKey.Get(t))
	})

	lastRequest := testcase.Let[*http.Request](s, func(t *testcase.T) *http.Request {
		return nil
	})
	next := testcase.Let(s, func(t *testcase.T) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			lastRequest.Set(t, r)
			w.WriteHeader(http.StatusTeapot)
		})
	})
	serviceName := testcase.Let(s, func(t *testcase.T) string {
		return t.Random.StringNC(5, random.CharsetASCII())
	})
	faultName := testcase.Let(s, func(t *testcase.T) string {
		return t.Random.StringNC(5, random.CharsetASCII())
	})
	newHandler := func(t *testcase.T, next http.Handler) http.Handler {
		return &fihttp.Handler{
			Next:        next,
			ServiceName: serviceName.Get(t),
			FaultsMapping: fihttp.HandlerFaultsMapping{
				faultName.Get(t): {faultKey{}},
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
				httpspec.Header.Get(t).Set(fihttp.HeaderName, string(data))
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
						t.Must.Equal(expectedErrOnFaultKey.Get(t), injector.Get(t).Check(lastRequest.Get(t).Context()))
					})
				})

				s.And("the injected fault name is not registered in the mapping", func(s *testcase.Spec) {
					name.Let(s, func(t *testcase.T) string {
						return faultName.Get(t) + t.Random.StringNC(5, random.CharsetAlpha())
					})

					s.Then("it will ignore the injected fault", func(t *testcase.T) {
						act(t)

						t.Must.NotNil(lastRequest.Get(t))
						t.Must.Nil(injector.Get(t).Check(lastRequest.Get(t).Context()))
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
							Next: httpspec.StubRoundTripper(func(r *http.Request) (*http.Response, error) {
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
					t.Must.Equal(1, len(lastRequest.Get(t).Header.Values(fihttp.HeaderName)))

					t.Logf("%#v", lastRequest.Get(t).Header.Values(fihttp.HeaderName))

					var faults []fihttp.Fault
					t.Must.Nil(json.Unmarshal([]byte(lastRequest.Get(t).Header.Get(fihttp.HeaderName)), &faults))
					t.Must.Contain(faults, injectedFaultInHeader.Get(t))
				})
			})
		})
	})
}

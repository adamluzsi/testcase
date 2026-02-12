package fihttp_test

import (
	"encoding/json"
	"net/http"
	"testing"

	"go.llib.dev/testcase/assert"
	"go.llib.dev/testcase/random"

	"go.llib.dev/testcase"
	"go.llib.dev/testcase/faultinject"
	"go.llib.dev/testcase/faultinject/fihttp"
	"go.llib.dev/testcase/tchttp"
)

func TestRoundTripper(t *testing.T) {
	s := testcase.NewSpec(t)
	faultinject.EnableForTest(t)

	var (
		next        = tchttp.LetRoundTripperRecorder(s)
		serviceName = testcase.LetValue(s, "")
	)
	newRoundTripper := func(t *testcase.T, next http.RoundTripper) http.RoundTripper {
		return &fihttp.RoundTripper{
			Next:        next,
			ServiceName: serviceName.Get(t),
		}
	}
	subject := testcase.Let(s, func(t *testcase.T) *fihttp.RoundTripper {
		return newRoundTripper(t, next.Get(t)).(*fihttp.RoundTripper)
	})

	s.Describe(".RoundTrip", func(s *testcase.Spec) {
		var request = tchttp.LetClientRequest(s, tchttp.RequestOption{})

		act := func(t *testcase.T) (*http.Response, error) {
			return subject.Get(t).RoundTrip(request.Get(t))
		}

		s.Context("it behaves as a http round tripper middleware",
			tchttp.RoundTripperMiddleware(newRoundTripper).Spec)

		s.When("propagated error is present", func(s *testcase.Spec) {
			fault := testcase.Let(s, func(t *testcase.T) fihttp.Fault {
				return fihttp.Fault{
					ServiceName: t.Random.StringNC(8, random.CharsetAlpha()),
					Name:        t.Random.StringNC(8, random.CharsetAlpha()),
				}
			})
			request.Let(s, func(t *testcase.T) *http.Request {
				super := request.Super(t)
				return super.WithContext(fihttp.Propagate(super.Context(), fault.Get(t)))
			})

			s.Then("outbound request will have the fault injection header", func(t *testcase.T) {
				_, err := act(t)
				assert.Must(t).NoError(err)
				lastRequest, ok := next.Get(t).LastReceivedRequest()
				assert.Must(t).True(ok, "expected that the request was received")
				header := lastRequest.Header.Get(fihttp.Header)
				assert.Must(t).NotEmpty(header)
				bytes, err := json.Marshal([]fihttp.Fault{fault.Get(t)})
				assert.Must(t).NoError(err)
				assert.Must(t).Contains(header, string(bytes))
			})
		})
	})
}

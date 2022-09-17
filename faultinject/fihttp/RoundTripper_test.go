package fihttp_test

import (
	"encoding/json"
	"github.com/adamluzsi/testcase/random"
	"net/http"
	"testing"

	"github.com/adamluzsi/testcase"
	"github.com/adamluzsi/testcase/faultinject"
	"github.com/adamluzsi/testcase/faultinject/fihttp"
	"github.com/adamluzsi/testcase/httpspec"
)

func TestRoundTripper(t *testing.T) {
	s := testcase.NewSpec(t)
	faultinject.EnableForTest(t)

	var (
		next        = httpspec.LetRoundTripperDouble(s)
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
		var (
			request = httpspec.OutboundRequest.Bind(s)
		)
		act := func(t *testcase.T) (*http.Response, error) {
			return subject.Get(t).RoundTrip(request.Get(t))
		}

		httpspec.ItBehavesLikeRoundTripperMiddleware(s, newRoundTripper)

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
				t.Must.NoError(err)
				header := next.Get(t).LastReceivedRequest(t).Header.Get(fihttp.Header)
				t.Must.NotEmpty(header)
				bytes, err := json.Marshal([]fihttp.Fault{fault.Get(t)})
				t.Must.NoError(err)
				t.Must.Contain(header, string(bytes))
			})
		})
	})
}

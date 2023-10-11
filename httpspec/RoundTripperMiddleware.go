package httpspec

import (
	"context"
	"io"
	"net/http"

	"github.com/adamluzsi/testcase"
)

func ItBehavesLikeRoundTripperMiddleware(s *testcase.Spec, subject MakeRoundTripperFunc) {
	testcase.RunSuite(s, RoundTripperMiddlewareContract{
		Subject: subject,
		MakeCTX: func(t *testcase.T) context.Context {
			return context.Background()
		},
	})
}

type RoundTripperMiddlewareContract struct {
	Subject MakeRoundTripperFunc
	MakeCTX testcase.VarInit[context.Context]
}

type MakeRoundTripperFunc func(t *testcase.T, next http.RoundTripper) http.RoundTripper

func (c RoundTripperMiddlewareContract) Spec(s *testcase.Spec) {
	s.Context(`it behaves like round-tripper`, func(s *testcase.Spec) {
		next := testcase.Let(s, func(t *testcase.T) *RoundTripperDouble {
			return &RoundTripperDouble{
				RoundTripperFunc: func(r *http.Request) (*http.Response, error) {
					return Response.Get(t), r.Context().Err()
				},
			}
		})
		subject := func(t *testcase.T) http.RoundTripper {
			return c.Subject(t, next.Get(t))
		}

		var (
			expectedBody = testcase.Let(s, func(t *testcase.T) string {
				return t.Random.String()
			})
			request = testcase.Let(s, func(t *testcase.T) *http.Request {
				r := OutboundRequest.Get(t)
				r.Body = asIOReader(t, r.Header, expectedBody.Get(t))
				return r
			})
		)
		act := func(t *testcase.T) (*http.Response, error) {
			return subject(t).RoundTrip(request.Get(t))
		}

		s.Test("round tripper act as a middleware in the round trip pipeline", func(t *testcase.T) {
			response, err := act(t)
			t.Must.Nil(err)

			// just some sanity check
			t.Must.Equal(Response.Get(t).StatusCode, response.StatusCode)
			t.Must.Equal(Response.Get(t).Status, response.Status)
			t.Must.ContainExactly(Response.Get(t).Header, response.Header)
		})

		s.Test("the next round tripper will receive the request", func(t *testcase.T) {
			_, err := act(t)
			t.Must.Nil(err)

			t.Must.Equal(1, len(next.Get(t).ReceivedRequests),
				"it should have received only one request")

			receivedRequest := next.Get(t).LastReceivedRequest(t)

			// just some sanity check
			t.Must.Equal(request.Get(t).URL.String(), receivedRequest.URL.String())
			t.Must.Equal(request.Get(t).Method, receivedRequest.Method)
			t.Must.ContainExactly(request.Get(t).Header, receivedRequest.Header)

			actualBody, err := io.ReadAll(receivedRequest.Body)
			t.Must.Nil(err)
			t.Must.Equal(expectedBody.Get(t), string(actualBody))
		})

		s.When("request context has an error", func(s *testcase.Spec) {
			Context := testcase.Let(s, func(t *testcase.T) context.Context {
				ctx, cancel := context.WithCancel(context.Background())
				cancel()
				return ctx
			})

			request.Let(s, func(t *testcase.T) *http.Request {
				return request.Super(t).WithContext(Context.Get(t))
			})

			s.Then("context error is propagated back", func(t *testcase.T) {
				_, err := act(t)
				t.Must.ErrorIs(err, Context.Get(t).Err())
			})
		})
	})
}

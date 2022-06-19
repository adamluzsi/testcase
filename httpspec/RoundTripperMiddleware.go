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
	MakeCTX testcase.VarInitFunc[context.Context]
}

type MakeRoundTripperFunc func(t *testcase.T, next http.RoundTripper) http.RoundTripper

func (c RoundTripperMiddlewareContract) Spec(s *testcase.Spec) {
	s.Context(`it behaves like round-tripper`, func(s *testcase.Spec) {
		var (
			receivedRequests = testcase.Let(s, func(t *testcase.T) []*http.Request {
				return []*http.Request{}
			})
			next = testcase.Let(s, func(t *testcase.T) StubRoundTripper {
				return func(r *http.Request) (*http.Response, error) {
					testcase.Append(t, receivedRequests, r)
					return Response.Get(t), nil
				}
			})
		)
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

			t.Must.Equal(1, len(receivedRequests.Get(t)), "it should have received a request")
			receivedRequest := receivedRequests.Get(t)[0]

			// just some sanity check
			t.Must.Equal(request.Get(t).URL.String(), receivedRequest.URL.String())
			t.Must.Equal(request.Get(t).Method, receivedRequest.Method)
			t.Must.ContainExactly(request.Get(t).Header, receivedRequest.Header)

			actualBody, err := io.ReadAll(receivedRequest.Body)
			t.Must.Nil(err)
			t.Must.Equal(expectedBody.Get(t), string(actualBody))
		})
	})
}

type StubRoundTripper func(r *http.Request) (*http.Response, error)

func (fn StubRoundTripper) RoundTrip(request *http.Request) (*http.Response, error) {
	return fn(request)
}

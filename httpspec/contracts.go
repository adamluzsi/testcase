package httpspec

import (
	"io"
	"net/http"

	"github.com/adamluzsi/testcase"
)

func ItBehavesLikeRoundTripper(s *testcase.Spec, subject func(t *testcase.T, next http.RoundTripper) http.RoundTripper) {
	testcase.RunContract(s, RoundTripperContract{Subject: subject})
}

type RoundTripperContract struct {
	Subject func(t *testcase.T, next http.RoundTripper) http.RoundTripper
}

func (c RoundTripperContract) Spec(s *testcase.Spec) {
	s.Context(`it behaves like round-tripper`, func(s *testcase.Spec) {
		expectedBody := testcase.Let(s, func(t *testcase.T) string {
			return t.Random.String()
		})
		Body.Let(s, func(t *testcase.T) any { return expectedBody.Get(t) })
		receivedRequests := testcase.Let(s, func(t *testcase.T) []*http.Request {
			return []*http.Request{}
		})
		next := testcase.Let(s, func(t *testcase.T) StubRoundTripper {
			return func(r *http.Request) (*http.Response, error) {
				testcase.Append(t, receivedRequests, r)
				return Response.Get(t), nil
			}
		})
		subject := func(t *testcase.T) http.RoundTripper {
			return c.Subject(t, next.Get(t))
		}
		act := func(t *testcase.T) (*http.Response, error) {
			return subject(t).RoundTrip(Request.Get(t))
		}

		s.Test("round tripper act as a middleware in the round trip pipeline", func(t *testcase.T) {
			response, err := act(t)
			t.Must.Nil(err)

			// just some sanity check
			t.Must.Equal(Response.Get(t).StatusCode, response.StatusCode)
			t.Must.Equal(Response.Get(t).Status, response.Status)
			t.Must.ContainExactly(Response.Get(t).Header, response.Header)
		})

		s.Test("the next round tripper receives the request", func(t *testcase.T) {
			_, err := act(t)
			t.Must.Nil(err)

			t.Must.Equal(1, len(receivedRequests.Get(t)), "it should have received a request")
			receivedRequest := receivedRequests.Get(t)[0]

			// just some sanity check
			t.Must.Equal(Request.Get(t).URL.String(), receivedRequest.URL.String())
			t.Must.Equal(Request.Get(t).Method, receivedRequest.Method)
			t.Must.ContainExactly(Request.Get(t).Header, receivedRequest.Header)

			actualBody, err := io.ReadAll(receivedRequest.Body)
			t.Must.Nil(err)
			t.Must.Equal(Body.Get(t).(string), string(actualBody))
		})
	})
}

type StubRoundTripper func(r *http.Request) (*http.Response, error)

func (fn StubRoundTripper) RoundTrip(request *http.Request) (*http.Response, error) {
	return fn(request)
}

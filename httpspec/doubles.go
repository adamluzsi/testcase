package httpspec

import (
	"bytes"
	"io"
	"net/http"
	"testing"

	"github.com/adamluzsi/testcase"
)

type RoundTripperFunc func(r *http.Request) (*http.Response, error)

func (fn RoundTripperFunc) RoundTrip(request *http.Request) (*http.Response, error) {
	return fn(request)
}

func LetRoundTripperDouble(s *testcase.Spec) testcase.Var[*RoundTripperDouble] {
	return testcase.Let(s, func(t *testcase.T) *RoundTripperDouble {
		return &RoundTripperDouble{}
	})
}

type RoundTripperDouble struct {
	// RoundTripperFunc is an optional argument in case you want to stub the response
	RoundTripperFunc RoundTripperFunc
	// ReceivedRequests hold all the received http request.
	ReceivedRequests []*http.Request
}

func (d *RoundTripperDouble) RoundTrip(r *http.Request) (*http.Response, error) {
	d.ReceivedRequests = append(d.ReceivedRequests, r.Clone(r.Context()))
	if d.RoundTripperFunc != nil {
		return d.RoundTripperFunc(r)
	}

	const code = http.StatusOK
	return &http.Response{
		Status:           http.StatusText(code),
		StatusCode:       code,
		Proto:            "HTTP/1.0",
		ProtoMajor:       1,
		ProtoMinor:       0,
		Header:           http.Header{},
		Body:             io.NopCloser(bytes.NewReader([]byte{})),
		ContentLength:    0,
		TransferEncoding: nil,
		Close:            false,
		Uncompressed:     false,
		Trailer:          nil,
		Request:          r,
		TLS:              nil,
	}, nil
}

func (d *RoundTripperDouble) LastReceivedRequest(tb testing.TB) *http.Request {
	if len(d.ReceivedRequests) == 0 {
		tb.Fatalf("%T did not received any *http.Request", *d)
	}
	return d.ReceivedRequests[len(d.ReceivedRequests)-1]
}

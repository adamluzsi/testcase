package tchttp

import (
	"bytes"
	"io"
	"net/http"

	"go.llib.dev/testcase"
)

type RoundTripperFunc func(r *http.Request) (*http.Response, error)

func (fn RoundTripperFunc) RoundTrip(request *http.Request) (*http.Response, error) {
	return fn(request)
}

func LetRoundTripperRecorder(s *testcase.Spec) testcase.Var[*RoundTripperRecorder] {
	return testcase.Let(s, func(t *testcase.T) *RoundTripperRecorder {
		return &RoundTripperRecorder{}
	})
}

type RoundTripperRecorder struct {
	// RoundTripperFunc is an optional argument in case you want to stub the response
	RoundTripperFunc RoundTripperFunc
	// ReceivedRequests hold all the received http request.
	ReceivedRequests []*http.Request
}

func (d *RoundTripperRecorder) RoundTrip(r *http.Request) (*http.Response, error) {
	d.ReceivedRequests = append(d.ReceivedRequests, r.Clone(r.Context()))
	if d.RoundTripperFunc != nil {
		return d.RoundTripperFunc(r)
	}
	if err := r.Context().Err(); err != nil {
		return nil, err
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

func (d *RoundTripperRecorder) LastReceivedRequest() (*http.Request, bool) {
	if len(d.ReceivedRequests) == 0 {
		return nil, false
	}
	return d.ReceivedRequests[len(d.ReceivedRequests)-1], true
}

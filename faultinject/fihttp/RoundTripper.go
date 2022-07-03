package fihttp

import (
	"encoding/json"
	"net/http"
	"syscall"
)

type (
	TagTimeout           struct{ ServiceName string }
	TagConnectionRefused struct{ ServiceName string }
)

type RoundTripper struct {
	Next        http.RoundTripper
	ServiceName string
}

func (rt RoundTripper) RoundTrip(r *http.Request) (*http.Response, error) {
	if _, ok := r.Context().Value(TagTimeout{ServiceName: rt.ServiceName}).(error); ok {
		return nil, netTimeoutError{}
	}

	if _, ok := r.Context().Value(TagConnectionRefused{ServiceName: rt.ServiceName}).(error); ok {
		return nil, syscall.ECONNREFUSED
	}

	if faults, ok := r.Context().Value(propagateCtxKey{}).([]Fault); ok {
		bs, err := json.Marshal(faults)
		if err != nil {
			return nil, err
		}
		r.Header.Set(HeaderName, string(bs))
	}

	return rt.Next.RoundTrip(r)
}

type netTimeoutError struct{}

func (e netTimeoutError) Error() string   { return "i/o timeout" }
func (e netTimeoutError) Timeout() bool   { return true }
func (e netTimeoutError) Temporary() bool { return true }

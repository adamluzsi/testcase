package fihttp

import (
	"encoding/json"
	"net/http"
)

type RoundTripper struct {
	Next http.RoundTripper
	// ServiceName is the name of the service of which this http.Client meant to do requests.
	ServiceName string
}

func (rt RoundTripper) RoundTrip(r *http.Request) (*http.Response, error) {
	if faults, ok := lookupFaults(r.Context()); ok {
		bs, err := json.Marshal(*faults)
		if err != nil {
			return nil, err
		}
		r.Header.Set(Header, string(bs))
	}
	return rt.Next.RoundTrip(r)
}

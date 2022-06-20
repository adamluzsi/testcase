package fihttp

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"syscall"

	"github.com/adamluzsi/testcase/faultinject"
)

type RoundTripper struct {
	Next        http.RoundTripper
	ServiceName string

	setUp     sync.Once
	injectori faultinject.Injector
}

func (rt *RoundTripper) init() {
	rt.setUp.Do(func() {
		rt.injectori = faultinject.Injector{}.
			OnTag(TagTimeout{ServiceName: rt.ServiceName}, netTimeoutError{}).
			OnTag(TagConnectionRefused{ServiceName: rt.ServiceName}, syscall.ECONNREFUSED)
	})
}

func (rt RoundTripper) RoundTrip(r *http.Request) (*http.Response, error) {
	rt.init()

	if err := rt.injectori.Check(r.Context()); err != nil {
		return nil, err
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

func servicePrefix(serviceName string) string {
	var prefix string
	if 0 < len(serviceName) {
		prefix = fmt.Sprintf("%s.", serviceName)
	}
	return prefix
}

type TagTimeout struct {
	ServiceName string
}

type TagConnectionRefused struct {
	ServiceName string
}

type netTimeoutError struct{}

func (e netTimeoutError) Error() string   { return "i/o timeout" }
func (e netTimeoutError) Timeout() bool   { return true }
func (e netTimeoutError) Temporary() bool { return true }

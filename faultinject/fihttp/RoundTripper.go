package fihttp

import (
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
			OnTag(TagTimeout(rt.ServiceName), netTimeoutError{}).
			OnTag(TagConnectionRefused(rt.ServiceName), syscall.ECONNREFUSED)
	})
}

func (rt RoundTripper) RoundTrip(r *http.Request) (*http.Response, error) {
	rt.init()

	if err := rt.injectori.Check(r.Context()); err != nil {
		return nil, err
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

func TagTimeout(serviceName string) string {
	return servicePrefix(serviceName) + "net.timeout"
}
func TagConnectionRefused(serviceName string) string {
	return servicePrefix(serviceName) + "net.connection-refused"
}

type netTimeoutError struct{}

func (e netTimeoutError) Error() string   { return "i/o timeout" }
func (e netTimeoutError) Timeout() bool   { return true }
func (e netTimeoutError) Temporary() bool { return true }

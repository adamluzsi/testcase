package fihttp_test

import (
	"context"
	"net/http"
	"strings"

	"github.com/adamluzsi/testcase/faultinject"
	"github.com/adamluzsi/testcase/faultinject/fihttp"
)

//func Example() {}

func ExampleRoundTripper() {
	const serviceName = "xy-service"
	c := &http.Client{
		Transport: fihttp.RoundTripper{
			Next:        http.DefaultTransport,
			ServiceName: serviceName,
		},
	}

	ctx := context.Background()

	// inject fault will make the client.Do fail with a timeout error once.
	// This is ideal if you want to test retry logic and such.
	ctx = faultinject.Inject(ctx, fihttp.TagTimeout(serviceName))

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "https://localhost:8080", strings.NewReader(""))
	if err != nil {
		panic(err)
	}

	response, err := c.Do(req)
	if err != nil {
		panic(err)
	}

	_ = response
}

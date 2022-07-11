package fihttp_test

import (
	"context"
	"errors"
	"log"
	"net/http"
	"strings"

	"github.com/adamluzsi/testcase/faultinject"
	"github.com/adamluzsi/testcase/faultinject/fihttp"
)

func Example() {
	type FaultTag struct{}

	client := &http.Client{
		Transport: fihttp.RoundTripper{
			Next:        http.DefaultTransport,
			ServiceName: "xy-external-service-name",
		},
	}

	myHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// if clients inject the "mapped-fault-name" then we will detect it here.
		if err := r.Context().Value(FaultTag{}).(error); err != nil {
			const code = http.StatusInternalServerError
			http.Error(w, http.StatusText(code), code)
			return
		}

		// outbound request will have faults injected which is not meant to our service
		outboundRequest, err := http.NewRequestWithContext(r.Context(), http.MethodGet, "http://example.com/", nil)
		if err != nil {
			const code = http.StatusInternalServerError
			http.Error(w, http.StatusText(code), code)
			return
		}
		_, _ = client.Do(outboundRequest)

		w.WriteHeader(http.StatusTeapot)
	})

	myHandlerWithFaultInjectionMiddleware := fihttp.Handler{
		Next:        myHandler,
		ServiceName: "our-service-name",
		FaultsMapping: fihttp.FaultsMapping{
			"mapped-fault-name": func(ctx context.Context) context.Context {
				return faultinject.Inject(ctx, FaultTag{}, errors.New("boom"))
			},
		},
	}

	if err := http.ListenAndServe(":8080", myHandlerWithFaultInjectionMiddleware); err != nil {
		log.Fatal(err.Error())
	}
}

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
	ctx = faultinject.Inject(ctx, fihttp.TagTimeout{ServiceName: serviceName}, nil)

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

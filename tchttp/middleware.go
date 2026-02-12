package tchttp

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"

	"go.llib.dev/testcase"
	"go.llib.dev/testcase/assert"
	"go.llib.dev/testcase/random"
)

type HandlerMiddlewareOption struct {
	Request testcase.VarInit[*http.Request]
}

func HandlerMiddleware(subject MakeHandlerMiddlewareFunc, opts ...HandlerMiddlewareOption) testcase.SpecSuite {
	s := testcase.NewSpec(nil)

	var c HandlerMiddlewareOption
	for _, opt := range opts {
		c.Request = cmpVarInitOr(opt.Request, c.Request)
	}

	var (
		LastReceivedRequest = testcase.Let(s, func(t *testcase.T) *http.Request {
			return nil
		})

		expectedResponseCode = testcase.Let(s, func(t *testcase.T) int {
			return t.Random.Pick([]int{
				http.StatusOK,
				http.StatusTeapot,
				http.StatusAccepted,
			}).(int)
		})
		expectedResponseHeader = testcase.Let(s, func(t *testcase.T) http.Header {
			var headers = http.Header{}
			t.Random.Repeat(1, 3, func() {
				key := fmt.Sprintf("X-%s", t.Random.StringNWithCharset(5, strings.ToUpper(random.CharsetAlpha())))
				value := t.Random.String()
				headers.Add(key, value)
			})
			return headers
		})
		expectedResponseBody = testcase.Let(s, func(t *testcase.T) string {
			return t.Random.String()
		})
		next = testcase.Let(s, func(t *testcase.T) http.HandlerFunc {
			return func(w http.ResponseWriter, r *http.Request) {
				LastReceivedRequest.Set(t, r)
				for k, v := range expectedResponseHeader.Get(t) {
					w.Header()[k] = v
				}
				w.WriteHeader(expectedResponseCode.Get(t))
				bs := []byte(expectedResponseBody.Get(t))
				n, err := w.Write(bs)
				assert.Must(t).Nil(err)
				assert.Must(t).Equal(len(bs), n)
			}
		})
	)

	makeMiddleware := func(t *testcase.T, next http.Handler) http.Handler {
		return subject(t, next)
	}
	middleware := testcase.Let(s, func(t *testcase.T) http.Handler {
		return makeMiddleware(t, next.Get(t))
	})

	var (
		recorder = testcase.Let(s, func(t *testcase.T) *httptest.ResponseRecorder {
			return httptest.NewRecorder()
		})

		expectedRequestBody = testcase.Let(s, func(t *testcase.T) []byte {
			var data = make([]byte, 1024)
			_, err := io.ReadFull(t.Random, data)
			assert.NoError(t, err)
			t.Random.Read(data)
			return data
		})

		request = testcase.Let(s, func(t *testcase.T) *http.Request {
			var req *http.Request
			if c.Request != nil {
				req = c.Request(t)
				data, err := io.ReadAll(req.Body)
				assert.NoError(t, err)
				assert.NoError(t, req.Body.Close())
				expectedRequestBody.Set(t, data)
				req.Body = io.NopCloser(bytes.NewReader(data))
			} else {
				req = defaultInboundHTTPRequestInit(t, bytes.NewReader(expectedRequestBody.Get(t)))
			}
			return req
		})
	)
	act := func(t *testcase.T) {
		middleware.Get(t).ServeHTTP(recorder.Get(t), request.Get(t))
	}

	s.Test("handler will propagate the request to the next http.Handler in the ServeHTTP pipeline", func(t *testcase.T) {
		act(t)

		assert.Must(t).NotNil(LastReceivedRequest.Get(t), "it was expected to receive a request in the next http.Handler")
		assert.Must(t).Equal(LastReceivedRequest.Get(t).Method, request.Get(t).Method)
		assert.Must(t).Equal(LastReceivedRequest.Get(t).URL.String(), request.Get(t).URL.String())
		assert.Must(t).Equal(LastReceivedRequest.Get(t).Header, request.Get(t).Header)

		data, err := io.ReadAll(LastReceivedRequest.Get(t).Body)
		assert.Must(t).Nil(err)
		assert.Must(t).Equal(expectedRequestBody.Get(t), []byte(data))

		// TODO: add more checks to ensure the last received request is functionally the same as the initial request.
	})

	s.Test("handler will propagate the ResponseWriter to the next http.Handler in the ServeHTTP pipeline", func(t *testcase.T) {
		act(t)

		rec := recorder.Get(t)
		assert.Must(t).Equal(expectedResponseCode.Get(t), rec.Code)
		assert.Must(t).Equal(expectedResponseBody.Get(t), rec.Body.String())
		for k, vs := range expectedResponseHeader.Get(t) {
			assert.Must(t).ContainsExactly(rec.Header()[k], vs)
		}
	})

	return s.AsSuite("http.Handler Middleware")
}

type MakeHandlerMiddlewareFunc func(t *testcase.T, next http.Handler) http.Handler

type MiddlewareConfig struct {
	Request  testcase.VarInit[*http.Request]
	Response testcase.VarInit[*http.Response]
}

func defaultInboundHTTPRequestInit(t *testcase.T, body io.Reader) *http.Request {
	t.Helper()
	req, err := http.NewRequest(http.MethodGet, "/", body)
	assert.NoError(t, err, "unexpected error with http.NewRequest")
	return req
}

type RoundTripperMiddlewareOption struct {
	Request  testcase.VarInit[*http.Request]
	Response testcase.VarInit[*http.Response]
}

func RoundTripperMiddleware(Subject RoundTripperMiddlewareFunc, opts ...RoundTripperMiddlewareOption) testcase.SpecSuite {
	s := testcase.NewSpec(nil)

	var c RoundTripperMiddlewareOption
	for _, opt := range opts {
		c.Request = cmpVarInitOr(opt.Request, c.Request)
		c.Response = cmpVarInitOr(opt.Response, c.Response)
	}

	var Response = testcase.Let(s, func(t *testcase.T) *http.Response {
		if c.Response != nil {
			return c.Response(t)
		}
		var (
			code = random.Pick(t.Random,
				http.StatusOK,
				http.StatusTeapot,
				http.StatusInternalServerError,
			)
			body = t.Random.String()
		)
		return &http.Response{
			Status:     http.StatusText(code),
			StatusCode: code,
			Proto:      "HTTP/1.0",
			ProtoMajor: 1,
			ProtoMinor: 0,
			Header: http.Header{
				"X-" + t.Random.StringNWithCharset(5, "ABCD"): {t.Random.StringNWithCharset(5, "ABCD")},
			},
			Body:          io.NopCloser(strings.NewReader(body)),
			ContentLength: int64(len(body)),
		}
	})

	next := testcase.Let(s, func(t *testcase.T) *RoundTripperRecorder {
		return &RoundTripperRecorder{
			RoundTripperFunc: func(r *http.Request) (*http.Response, error) {
				return Response.Get(t), r.Context().Err()
			},
		}
	})
	subject := func(t *testcase.T) http.RoundTripper {
		return Subject(t, next.Get(t))
	}

	var (
		expectedBody = testcase.Let(s, func(t *testcase.T) []byte {
			var data = make([]byte, 1024)
			_, err := io.ReadFull(t.Random, data)
			assert.NoError(t, err)
			t.Random.Read(data)
			return data
		})
		request = testcase.Let[*http.Request](s, func(t *testcase.T) *http.Request {
			var req *http.Request
			if c.Request != nil {
				req = c.Request(t)
				data, err := io.ReadAll(req.Body)
				assert.NoError(t, err)
				assert.NoError(t, req.Body.Close())
				expectedBody.Set(t, data)
				req.Body = io.NopCloser(bytes.NewReader(data))
			} else {
				req = defaultOutbountHTTPRequest(t)
				req.Body = io.NopCloser(bytes.NewReader([]byte(expectedBody.Get(t))))
			}
			return req
		})
	)
	act := func(t *testcase.T) (*http.Response, error) {
		return subject(t).RoundTrip(request.Get(t))
	}

	s.Test("round tripper act as a middleware in the round trip pipeline", func(t *testcase.T) {
		response, err := act(t)
		assert.Must(t).Nil(err)

		// just some sanity check
		assert.Must(t).Equal(Response.Get(t).StatusCode, response.StatusCode)
		assert.Must(t).Equal(Response.Get(t).Status, response.Status)
		assert.Must(t).ContainsExactly(Response.Get(t).Header, response.Header)
	})

	s.Test("the next round tripper will receive the request", func(t *testcase.T) {
		_, err := act(t)
		assert.Must(t).Nil(err)

		assert.Must(t).Equal(1, len(next.Get(t).ReceivedRequests),
			"it should have received only one request")

		receivedRequest, ok := next.Get(t).LastReceivedRequest()
		assert.Must(t).True(ok, "expected that final round tripper received the outbound http request")

		// just some sanity check
		assert.Must(t).Equal(request.Get(t).URL.String(), receivedRequest.URL.String())
		assert.Must(t).Equal(request.Get(t).Method, receivedRequest.Method)
		assert.Must(t).ContainsExactly(request.Get(t).Header, receivedRequest.Header)

		actualBody, err := io.ReadAll(receivedRequest.Body)
		assert.Must(t).Nil(err)
		assert.Must(t).Equal(expectedBody.Get(t), actualBody)
	})

	s.When("request context has an error", func(s *testcase.Spec) {
		Context := testcase.Let(s, func(t *testcase.T) context.Context {
			ctx, cancel := context.WithCancel(context.Background())
			cancel()
			return ctx
		})

		request.Let(s, func(t *testcase.T) *http.Request {
			return request.Super(t).WithContext(Context.Get(t))
		})

		s.Then("context error is propagated back", func(t *testcase.T) {
			_, err := act(t)
			assert.Must(t).ErrorIs(err, Context.Get(t).Err())
		})
	})

	return s.AsSuite("http.RoundTripper middleware")
}

type RoundTripperMiddlewareFunc func(t *testcase.T, next http.RoundTripper) http.RoundTripper

func defaultOutbountHTTPRequest(t *testcase.T) *http.Request {
	u := url.URL{
		Scheme: t.Random.Pick([]string{"http", "https"}).(string),
		Host:   fmt.Sprintf("www.%s.com", t.Random.StringNC(7, random.CharsetAlpha())),
		Path:   "/",
	}
	r, err := http.NewRequest(http.MethodGet, u.String(), nil)
	assert.Must(t).Nil(err)
	return r
}

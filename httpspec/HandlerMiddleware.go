package httpspec

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"

	"github.com/adamluzsi/testcase"
)

func ItBehavesLikeHandlerMiddleware(s *testcase.Spec, subject MakeHandlerMiddlewareFunc) {
	testcase.RunSuite(s, HandlerMiddlewareContract{
		Subject: subject,
		MakeCTX: func(t *testcase.T) context.Context {
			return context.Background()
		},
	})
}

type HandlerMiddlewareContract struct {
	Subject MakeHandlerMiddlewareFunc
	MakeCTX testcase.VarInit[context.Context]
}

type MakeHandlerMiddlewareFunc func(t *testcase.T, next http.Handler) http.Handler

func (c HandlerMiddlewareContract) Spec(s *testcase.Spec) {
	s.Context(`it behaves like http.Handler Middleware`, func(s *testcase.Spec) {
		var (
			LastReceivedRequest = testcase.Let(s, func(t *testcase.T) *http.Request { return nil })

			expectedResponseCode = testcase.Let(s, func(t *testcase.T) int {
				return t.Random.ElementFromSlice([]int{
					http.StatusOK,
					http.StatusTeapot,
					http.StatusAccepted,
				}).(int)
			})
			expectedResponseHeader = testcase.Let(s, func(t *testcase.T) http.Header {
				return http.Header{
					"foo": {t.Random.String()},
				}
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
					t.Must.Nil(err)
					t.Must.Equal(len(bs), n)
				}
			})
		)

		makeMiddleware := func(t *testcase.T, next http.Handler) http.Handler {
			return c.Subject(t, next)
		}
		middleware := testcase.Let(s, func(t *testcase.T) http.Handler {
			return makeMiddleware(t, next.Get(t))
		})

		var (
			recorder = testcase.Let(s, func(t *testcase.T) *httptest.ResponseRecorder {
				return ResponseRecorder.Get(t)
			})
			expReqBody = testcase.Let(s, func(t *testcase.T) string {
				return t.Random.String()
			})
			request = testcase.Let(s, func(t *testcase.T) *http.Request {
				r := InboundRequest.Get(t)
				r.Body = asIOReader(t, r.Header, expReqBody.Get(t))
				return r
			})
		)
		act := func(t *testcase.T) {
			middleware.Get(t).ServeHTTP(recorder.Get(t), request.Get(t))
		}

		s.Test("handler will propagate the request to the next http.Handler in the ServeHTTP pipeline", func(t *testcase.T) {
			act(t)

			t.Must.NotNil(LastReceivedRequest.Get(t), "it was expected to receive a request in the next http.Handler")
			t.Must.Equal(LastReceivedRequest.Get(t).Method, request.Get(t).Method)
			t.Must.Equal(LastReceivedRequest.Get(t).URL.String(), request.Get(t).URL.String())
			t.Must.Equal(LastReceivedRequest.Get(t).Header, request.Get(t).Header)

			data, err := io.ReadAll(LastReceivedRequest.Get(t).Body)
			t.Must.Nil(err)
			t.Must.Equal(expReqBody.Get(t), string(data))

			// TODO: add more checks to ensure the last received request is functionally the same as the initial request.
		})

		s.Test("handler will propagate the ResponseWriter to the next http.Handler in the ServeHTTP pipeline", func(t *testcase.T) {
			act(t)

			rec := recorder.Get(t)
			t.Must.Equal(expectedResponseCode.Get(t), rec.Code)
			t.Must.Equal(expectedResponseBody.Get(t), rec.Body.String())
			for k, vs := range expectedResponseHeader.Get(t) {
				t.Must.ContainExactly(rec.Header()[k], vs)
			}
		})
	})
}

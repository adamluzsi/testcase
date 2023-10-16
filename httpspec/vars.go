package httpspec

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"

	"go.llib.dev/testcase"
	"go.llib.dev/testcase/random"
)

var (
	// Handler prepares the current testcase spec scope to be ready for http handler testing.
	// You define your spec subject with this and all the request will be pointed towards this.
	Handler = testcase.Var[http.Handler]{ID: `httpspec:Handler`}
	// Context allow to retrieve the current test scope's request context.
	Context = testcase.Var[context.Context]{ID: `httpspec:Context`, Init: func(t *testcase.T) context.Context {
		return context.Background()
	}}
	Method = testcase.Var[string]{ID: `httpspec:Method`, Init: func(t *testcase.T) string {
		return http.MethodGet
	}}
	Path = testcase.Var[string]{ID: `httpspec:Path`, Init: func(t *testcase.T) string {
		return `/`
	}}
	// Query allows you to retrieve the current test scope's http PathGet query that will be used for ServeHTTP.
	// In a Before Block you can access the query and then specify the values in it.
	Query = testcase.Var[url.Values]{ID: `httpspec:Query`, Init: func(t *testcase.T) url.Values {
		return url.Values{}
	}}
	// Header allows you to set the current test scope's http PathGet for ServeHTTP.
	Header = testcase.Var[http.Header]{ID: `httpspec:Header.Get`, Init: func(t *testcase.T) http.Header {
		return http.Header{}
	}}
)

var (
	InboundRequest = testcase.Var[*http.Request]{
		ID: "httpspec:InboundRequest",
		Init: func(t *testcase.T) *http.Request {
			target, _ := url.Parse(Path.Get(t))
			target.RawQuery = Query.Get(t).Encode()
			if isDebugEnabled(t) {
				t.Log(`http.InboundRequest.Method:`, Method.Get(t))
				t.Log(`http.InboundRequest.Path`, target.String())
			}
			r := httptest.NewRequest(Method.Get(t), target.String(), asIOReader(t, Header.Get(t), Body.Get(t)))
			r = r.WithContext(Context.Get(t))
			r.Header = Header.Get(t)
			return r
		},
	}
	OutboundRequest = testcase.Var[*http.Request]{
		ID: "httpspec:OutboundRequest",
		Init: func(t *testcase.T) *http.Request {
			u := url.URL{
				Scheme:   t.Random.SliceElement([]string{"http", "https"}).(string),
				Host:     fmt.Sprintf("www.%s.com", t.Random.StringNC(7, random.CharsetAlpha())),
				Path:     Path.Get(t),
				RawPath:  Path.Get(t),
				RawQuery: Query.Get(t).Encode(),
			}
			if isDebugEnabled(t) {
				t.Log(`http.OutboundRequest.Method:`, Method.Get(t))
				t.Log(`http.OutboundRequest.Path`, u.Path)
			}
			r, err := http.NewRequest(Method.Get(t), u.String(), asIOReader(t, Header.Get(t), Body.Get(t)))
			t.Must.Nil(err)
			r = r.WithContext(Context.Get(t))
			r.Header = Header.Get(t)
			return r
		},
	}
	ResponseRecorder = testcase.Var[*httptest.ResponseRecorder]{
		ID: "httpspec:ResponseRecorder",
		Init: func(t *testcase.T) *httptest.ResponseRecorder {
			return httptest.NewRecorder()
		},
	}
	Response = testcase.Var[*http.Response]{
		ID: "httpspec:Response",
		Init: func(t *testcase.T) *http.Response {
			code := t.Random.SliceElement([]int{
				http.StatusOK,
				http.StatusTeapot,
				http.StatusInternalServerError,
			}).(int)
			body := t.Random.String()
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
		},
	}
)

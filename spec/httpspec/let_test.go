package httpspec_test

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"path"
	"testing"

	"go.llib.dev/testcase"
	"go.llib.dev/testcase/assert"
	"go.llib.dev/testcase/random"
	"go.llib.dev/testcase/spec/httpspec"
)

func ExampleLetResponseRecorder() {
	s := testcase.NewSpec(nil)

	rr := httpspec.LetResponseRecorder(s)

	s.Test("", func(t *testcase.T) {
		_ = rr.Get(t)
	})
}

func TestLetResponseRecorder(t *testing.T) {
	s := testcase.NewSpec(t)
	rr := httpspec.LetResponseRecorder(s)
	s.Test("", func(t *testcase.T) {
		t.Must.Empty(rr.Get(t).Body.String())
		_, err := rr.Get(t).WriteString("hello")
		t.Must.NoError(err)
		t.Must.Contains(rr.Get(t).Body.String(), "hello")
	})
}

func ExampleLetClientRequest() {
	s := testcase.NewSpec(nil)

	request := httpspec.LetClientRequest(s, httpspec.RequestVar{})

	s.Test("", func(t *testcase.T) {
		_ = request.Get(t)
	})
}

func TestLetClientRequest(t *testing.T) {
	s := testcase.NewSpec(t)

	s.When("RequestVar is empty", func(s *testcase.Spec) {
		request := httpspec.LetClientRequest(s, httpspec.RequestVar{})
		s.Then("default values used", func(t *testcase.T) {
			r := request.Get(t)
			t.Must.NotEmpty(r.Host)
			t.Must.Equal(http.MethodGet, r.Method)
			t.Must.Equal("/", r.URL.Path)
			t.Must.Empty(r.URL.Query())
			t.Must.Empty(r.Header)
			t.Must.Empty(assert.ReadAll(t, r.Body))
		})
	})

	s.When("RequestVar is populated with values", func(s *testcase.Spec) {
		type BodyDTO struct {
			V1 string `json:"v1"`
			V2 int    `json:"v2"`
		}
		rv := httpspec.RequestVar{
			Context: testcase.Let(s, func(t *testcase.T) context.Context {
				return context.WithValue(context.Background(), "foo", "bar")
			}),
			Scheme: testcase.Let(s, func(t *testcase.T) string {
				return "postgres"
			}),
			Host: testcase.Let(s, func(t *testcase.T) string {
				return fmt.Sprintf("www.%s.com", t.Random.StringNC(5, random.CharsetAlpha()))
			}),
			Method: testcase.Let(s, func(t *testcase.T) string {
				return t.Random.Pick([]string{
					http.MethodGet,
					http.MethodPost,
					http.MethodPut,
					http.MethodDelete,
				}).(string)
			}),
			Path: testcase.Let(s, func(t *testcase.T) string {
				cs := random.CharsetAlpha()
				return "/" + path.Join(t.Random.StringNC(3, cs), t.Random.StringNC(5, cs))
			}),
			Query: testcase.Let(s, func(t *testcase.T) url.Values {
				return url.Values{t.Random.StringNC(5, random.CharsetAlpha()): []string{t.Random.StringN(5)}}
			}),
			Header: testcase.Let(s, func(t *testcase.T) http.Header {
				h := http.Header{}
				charset := random.CharsetAlpha()
				h.Set(t.Random.StringNC(5, charset), t.Random.StringNC(5, charset))
				h.Set("Content-Type", "application/json")
				return h
			}),
			Body: testcase.Let(s, func(t *testcase.T) any {
				return BodyDTO{
					V1: t.Random.String(),
					V2: t.Random.Int(),
				}
			}),
		}
		request := httpspec.LetClientRequest(s, rv)

		s.Test("injected variables used", func(t *testcase.T) {
			r := request.Get(t)
			t.Must.Equal(rv.Header.Get(t), r.Header)
			t.Must.Equal(rv.Path.Get(t), r.URL.Path)
			t.Must.Equal(rv.Query.Get(t), r.URL.Query())
			t.Must.Equal(rv.Scheme.Get(t), r.URL.Scheme)
			t.Must.Equal(rv.Method.Get(t), r.Method)
			t.Must.Equal(rv.Context.Get(t), r.Context())
			var body BodyDTO
			t.Must.NoError(json.Unmarshal(t.Must.ReadAll(r.Body), &body))
			t.Must.Equal(rv.Body.Get(t), body)
		})
	})
}

func ExampleLetServerRequest() {
	s := testcase.NewSpec(nil)

	request := httpspec.LetServerRequest(s, httpspec.RequestVar{})

	s.Test("", func(t *testcase.T) {
		_ = request.Get(t)
	})
}

func TestLetServerRequest(t *testing.T) {
	s := testcase.NewSpec(t)

	s.When("RequestVar is empty", func(s *testcase.Spec) {
		request := httpspec.LetServerRequest(s, httpspec.RequestVar{})

		s.Then("default values used", func(t *testcase.T) {
			r := request.Get(t)
			t.Must.NotEmpty(r.Host)
			t.Must.Equal(http.MethodGet, r.Method)
			t.Must.Equal("/", r.URL.Path)
			t.Must.Empty(r.URL.Query())
			t.Must.Empty(r.Header)
			t.Must.Empty(assert.ReadAll(t, r.Body))
		})

		s.Then("is a server request", func(t *testcase.T) {
			r := request.Get(t)
			assert.NotEmpty(t, r.RemoteAddr)
			t.Must.NotEmpty(r.Host)
			// For HTTPS URLs, TLS is non-nil
			if r.URL.Scheme == "https" {
				t.Must.NotNil(r.TLS)
			}
		})
	})

	s.When("RequestVar is populated with values", func(s *testcase.Spec) {
		type BodyDTO struct {
			V1 string `json:"v1"`
			V2 int    `json:"v2"`
		}
		rv := httpspec.RequestVar{
			Context: testcase.Let(s, func(t *testcase.T) context.Context {
				return context.WithValue(context.Background(), "foo", "bar")
			}),
			Scheme: testcase.Let(s, func(t *testcase.T) string {
				return "postgres"
			}),
			Host: testcase.Let(s, func(t *testcase.T) string {
				return fmt.Sprintf("www.%s.com", t.Random.StringNC(5, random.CharsetAlpha()))
			}),
			Method: testcase.Let(s, func(t *testcase.T) string {
				return t.Random.Pick([]string{
					http.MethodGet,
					http.MethodPost,
					http.MethodPut,
					http.MethodDelete,
				}).(string)
			}),
			Path: testcase.Let(s, func(t *testcase.T) string {
				cs := random.CharsetAlpha()
				return "/" + path.Join(t.Random.StringNC(3, cs), t.Random.StringNC(5, cs))
			}),
			Query: testcase.Let(s, func(t *testcase.T) url.Values {
				return url.Values{t.Random.StringNC(5, random.CharsetAlpha()): []string{t.Random.StringN(5)}}
			}),
			Header: testcase.Let(s, func(t *testcase.T) http.Header {
				h := http.Header{}
				charset := random.CharsetAlpha()
				h.Set(t.Random.StringNC(5, charset), t.Random.StringNC(5, charset))
				h.Set("Content-Type", "application/json")
				return h
			}),
			Body: testcase.Let(s, func(t *testcase.T) any {
				return BodyDTO{
					V1: t.Random.String(),
					V2: t.Random.Int(),
				}
			}),
		}
		request := httpspec.LetClientRequest(s, rv)

		s.Test("injected variables used", func(t *testcase.T) {
			r := request.Get(t)
			t.Must.Equal(rv.Header.Get(t), r.Header)
			t.Must.Equal(rv.Path.Get(t), r.URL.Path)
			t.Must.Equal(rv.Query.Get(t), r.URL.Query())
			t.Must.Equal(rv.Scheme.Get(t), r.URL.Scheme)
			t.Must.Equal(rv.Method.Get(t), r.Method)
			t.Must.Equal(rv.Context.Get(t), r.Context())
			var body BodyDTO
			t.Must.NoError(json.Unmarshal(t.Must.ReadAll(r.Body), &body))
			t.Must.Equal(rv.Body.Get(t), body)
		})
	})
}

func TestLetServer(t *testing.T) {
	s := testcase.NewSpec(t)
	s.HasSideEffect()
	s.Sequential()

	srv := httpspec.LetServer(s, func(t *testcase.T) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusTeapot)
		})
	})

	var leak *httptest.Server
	s.Test("", func(t *testcase.T) {
		response, err := srv.Get(t).Client().Get(srv.Get(t).URL)
		t.Must.NoError(err)
		t.Must.Equal(http.StatusTeapot, response.StatusCode)
		leak = srv.Get(t)
	})

	s.Finish()
	_, err := leak.Client().Get(leak.URL)
	assert.NotNil(t, err, "should be closed after the test")
}

func TestServerClientDo(t *testing.T) {
	s := testcase.NewSpec(t)

	req := httpspec.LetClientRequest(s, httpspec.RequestVar{
		Path: testcase.Let(s, func(t *testcase.T) string {
			return "/" + url.PathEscape(t.Random.String())
		}),
		Query: testcase.Let(s, func(t *testcase.T) url.Values {
			q := url.Values{}
			q.Set("foo", t.Random.String())
			return q

		}),
		Header: testcase.Let(s, func(t *testcase.T) http.Header {
			h := http.Header{}
			h.Set("bar", "baz")
			return h
		}),
	})
	srv := httpspec.LetServer(s, func(t *testcase.T) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, actual *http.Request) {
			expected := req.Get(t)
			t.Should.Equal(expected.URL.Path, actual.URL.Path)
			t.Should.Equal(expected.URL.Query(), actual.URL.Query())
			t.Should.Equal(expected.URL.Query(), actual.URL.Query())
			w.WriteHeader(http.StatusTeapot)
		})
	})

	s.Test("", func(t *testcase.T) {
		response, err := httpspec.ServerClientDo(t, srv.Get(t), req.Get(t))
		t.Must.NoError(err)
		t.Must.Equal(http.StatusTeapot, response.StatusCode)
	})
}

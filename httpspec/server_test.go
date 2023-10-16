package httpspec_test

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"go.llib.dev/testcase"
	"go.llib.dev/testcase/assert"
	"go.llib.dev/testcase/httpspec"
)

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

func TestClientDo(t *testing.T) {
	s := testcase.NewSpec(t)

	req := httpspec.LetRequest(s, httpspec.RequestVar{
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
		response, err := httpspec.ClientDo(t, srv.Get(t), req.Get(t))
		t.Must.NoError(err)
		t.Must.Equal(http.StatusTeapot, response.StatusCode)
	})
}

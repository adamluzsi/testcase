package httpspec_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/adamluzsi/testcase"
	"github.com/adamluzsi/testcase/assert"
	"github.com/adamluzsi/testcase/httpspec"
)

func TestLetResponseRecorder(t *testing.T) {
	s := testcase.NewSpec(t)
	rr := httpspec.LetResponseRecorder(s)
	s.Test("", func(t *testcase.T) {
		t.Must.Empty(rr.Get(t).Body.String())
		_, err := rr.Get(t).WriteString("hello")
		t.Must.NoError(err)
		t.Must.Contain(rr.Get(t).Body.String(), "hello")
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

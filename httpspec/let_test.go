package httpspec_test

import (
	"net/http"
	"testing"

	"github.com/adamluzsi/testcase"
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

func TestLetInboundRequest(t *testing.T) {
	s := testcase.NewSpec(t)
	req := httpspec.LetInboundRequest(s)
	s.Test("", func(t *testcase.T) {
		t.Must.Equal(http.MethodGet, req.Get(t).Method)
		t.Must.Equal("/", req.Get(t).URL.String())
	})
}

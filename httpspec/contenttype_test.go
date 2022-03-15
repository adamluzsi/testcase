package httpspec_test

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"testing"

	"github.com/adamluzsi/testcase"
	"github.com/adamluzsi/testcase/httpspec"
)

func TestContentTypeIsJSON(t *testing.T) {
	s := testcase.NewSpec(t)

	var actually map[string]string
	httpspec.Handler.Let(s, func(t *testcase.T) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer r.Body.Close()
			t.Must.Equal(`application/json`, r.Header.Get(`Content-Type`))
			bs, err := ioutil.ReadAll(r.Body)
			t.Must.Nil(err)
			t.Must.Nil(json.Unmarshal(bs, &actually))
		})
	})

	httpspec.ContentTypeIsJSON(s)

	expected := map[string]string{"hello": "world"}
	httpspec.Body.Let(s, func(t *testcase.T) interface{} { return expected })

	s.Test(`test json encoding for actually`, func(t *testcase.T) {
		httpspec.ServeHTTP(t)

		t.Must.Equal(expected, actually)
	})
}

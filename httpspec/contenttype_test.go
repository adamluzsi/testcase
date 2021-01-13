package httpspec_test

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/adamluzsi/testcase"
	"github.com/adamluzsi/testcase/httpspec"
)

func TestContentTypeIsJSON(t *testing.T) {
	s := testcase.NewSpec(t)

	var actually map[string]string
	httpspec.SubjectLet(s, func(t *testcase.T) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer r.Body.Close()
			require.Equal(t, `application/json`, r.Header.Get(`Content-Type`))
			bs, err := ioutil.ReadAll(r.Body)
			require.Nil(t, err)
			require.Nil(t, json.Unmarshal(bs, &actually))
		})
	})

	httpspec.ContentTypeIsJSON(s)

	expected := map[string]string{"hello": "world"}
	httpspec.Body.Let(s, func(t *testcase.T) interface{} { return expected })

	s.Test(`test json encoding for actually`, func(t *testcase.T) {
		httpspec.SubjectGet(t)

		require.Equal(t, expected, actually)
	})
}

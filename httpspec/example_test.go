package httpspec_test

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/adamluzsi/testcase"
	"github.com/adamluzsi/testcase/httpspec"
)

func Example_usage() {
	var tb testing.TB
	s := testcase.NewSpec(tb)

	// subject
	httpspec.Handler.Let(s, func(t *testcase.T) http.Handler {
		return MyHandler{}
	})

	// Arrange
	httpspec.ContentTypeIsJSON(s)
	httpspec.Method.LetValue(s, http.MethodPost)
	httpspec.Path.LetValue(s, `/`)
	httpspec.Body.Let(s, func(t *testcase.T) interface{} {
		// this will end up as {"foo":"bar"} in the request body
		return map[string]string{"foo": "bar"}
	})

	s.Then(`it will...`, func(t *testcase.T) {
		// ServeHTTP
		rr := httpspec.ServeHTTP(t)

		// Assert
		t.Must.Equal(http.StatusOK, rr.Code)
		var resp CreateResponse
		t.Must.Nil(json.Unmarshal(rr.Body.Bytes(), &resp))
		// ...
	})
}

package httpspec_test

import (
	"encoding/json"
	"github.com/adamluzsi/testcase"
	"github.com/adamluzsi/testcase/httpspec"
	"github.com/stretchr/testify/require"
	"net/http"
	"testing"
)

func Example_usage() {
	var tb testing.TB
	s := testcase.NewSpec(tb)

	// subject
	httpspec.SubjectLet(s, func(t *testcase.T) http.Handler {
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
		// Act
		rr := httpspec.SubjectGet(t)

		// Assert
		require.Equal(t, http.StatusOK, rr.Code)
		var resp CreateResponse
		require.Nil(t, json.Unmarshal(rr.Body.Bytes(), &resp))
		// ...
	})
}

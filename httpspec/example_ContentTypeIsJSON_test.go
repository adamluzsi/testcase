package httpspec_test

import (
	"encoding/json"
	"net/http"

	"go.llib.dev/testcase"
	"go.llib.dev/testcase/assert"

	. "go.llib.dev/testcase/httpspec"
)

func ExampleContentTypeIsJSON() {
	s := testcase.NewSpec(testingT)

	Handler.Let(s, func(t *testcase.T) http.Handler { return MyHandler{} })
	ContentTypeIsJSON(s)

	s.Describe(`POST / - create X`, func(s *testcase.Spec) {
		Method.LetValue(s, http.MethodPost)
		Path.LetValue(s, `/`)

		Body.Let(s, func(t *testcase.T) interface{} {
			// this will end up as {"foo":"bar"} in the request body
			return map[string]string{"foo": "bar"}
		})

		var onSuccess = func(t *testcase.T) CreateResponse {
			rr := ServeHTTP(t)
			assert.Must(t).Equal(http.StatusOK, rr.Code)
			var resp CreateResponse
			assert.Must(t).Nil(json.Unmarshal(rr.Body.Bytes(), &resp))
			return resp
		}

		s.Then(`it will create a new resource`, func(t *testcase.T) {
			createResponse := onSuccess(t)
			// assert
			_ = createResponse
		})
	})
}

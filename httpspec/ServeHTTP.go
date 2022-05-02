package httpspec

import (
	"net/http/httptest"

	"github.com/adamluzsi/testcase"
)

var serveOnce = testcase.Var[struct{}]{
	ID: "httpspec:ServeHTTP",
	Init: func(t *testcase.T) struct{} {
		Handler.Get(t).ServeHTTP(ResponseRecorder.Get(t), Request.Get(t))
		return struct{}{}
	},
}

// ServeHTTP will make a request to the spec context
// it requires the following spec variables
//	* Method -> http MethodGet <string>
//	* Path -> http PathGet <string>
//	* Query -> http query string <url.Values>
//	* Body -> http payload <io.Reader|io.ReadCloser>
//
func ServeHTTP(t *testcase.T) *httptest.ResponseRecorder {
	serveOnce.Get(t)
	return ResponseRecorder.Get(t)
}

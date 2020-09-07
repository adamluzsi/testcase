package httpspec

import (
	"context"
	"net/http"
	"net/url"

	"github.com/adamluzsi/testcase"
)

const (
	letVarPrefix          = `httpspec:`
	ContextVarName        = letVarPrefix + `context`
	requestMethodVarName  = letVarPrefix + `method`
	requestPathVarName    = letVarPrefix + `path`
	requestQueryVarName   = letVarPrefix + `query`
	requestHeaderVarName  = letVarPrefix + `header`
	requestBodyVarName    = letVarPrefix + `body`
	requestHandlerVarName = letVarPrefix + `handler`
)

// LetContext allow you to Set the ServeHTTP request context
func LetContext(s *testcase.Spec, f func(t *testcase.T) context.Context) {
	s.Let(ContextVarName, func(t *testcase.T) interface{} { return f(t) })
}

func ctx(t *testcase.T) context.Context {
	return t.I(ContextVarName).(context.Context)
}

// LetMethod allow you to set the current test scope's http method for ServeHTTP
func LetMethod(s *testcase.Spec, f func(t *testcase.T) string) {
	s.Let(requestMethodVarName, func(t *testcase.T) interface{} { return f(t) })
}

// LetMethodValue allow you to set the current test scope's http method for ServeHTTP
func LetMethodValue(s *testcase.Spec, m string) {
	s.LetValue(requestMethodVarName, m)
}

// method returns you the currently defined http method Value that will be used for ServeHTTP
func method(t *testcase.T) string {
	return t.I(requestMethodVarName).(string)
}

// LetPath allows you to set the current test scope's http path for ServeHTTP.
func LetPath(s *testcase.Spec, f func(t *testcase.T) string) {
	s.Let(requestPathVarName, func(t *testcase.T) interface{} { return f(t) })
}

// LetPathValue allows you to set the current test scope's http path for ServeHTTP.
func LetPathValue(s *testcase.Spec, p string) {
	s.LetValue(requestPathVarName, p)
}

// path returns the current test scope's http path that will be used for the ServeHTTP.
// The Query string part is not part of the path definition here.
func path(t *testcase.T) string {
	return t.I(requestPathVarName).(string)
}

// letQuery allows you to set the current test scope's http path query for ServeHTTP.
// It is advised to use Query instead of letQuery to incrementally build up the query string content for the request,
// instead of overriding the whole in the given scope.
func letQuery(s *testcase.Spec, f func(t *testcase.T) url.Values) {
	s.Let(requestQueryVarName, func(t *testcase.T) interface{} { return f(t) })
}

// Query allows you to retrieve the current test scope's http path query that will be used for ServeHTTP.
// In a Before Block you can access the query and then specify the values in it.
func Query(t *testcase.T) url.Values {
	return t.I(requestQueryVarName).(url.Values)
}

// letHandler allows you to set the current test scope's http header for ServeHTTP.
func letHeader(s *testcase.Spec, f func(t *testcase.T) http.Header) {
	s.Let(requestHeaderVarName, func(t *testcase.T) interface{} { return f(t) })
}

// Header allows you to set the current test scope's http path for ServeHTTP.
func Header(t *testcase.T) http.Header {
	return t.I(requestHeaderVarName).(http.Header)
}

// LetBody allow you to define a http request body value for the ServeHTTP.
// The value of this can be a struct, map or url.Values.
// The Serialization for the request body is based on the Header "Content-Type" value.
// Currently only json and form url encoding is supported.
func LetBody(s *testcase.Spec, f func(t *testcase.T) interface{}) {
	s.Let(requestBodyVarName, f)
}

// body returns the defined body object
func body(t *testcase.T) interface{} {
	return t.I(requestBodyVarName)
}

// letHandler is the subject of a HTTP Spec.
func letHandler(s *testcase.Spec, f func(t *testcase.T) http.Handler) {
	s.Let(requestHandlerVarName, func(t *testcase.T) interface{} { return f(t) })
}

// handler returns the current test scope's http.handler.
func handler(t *testcase.T) http.Handler {
	return t.I(requestHandlerVarName).(http.Handler)
}

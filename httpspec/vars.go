package httpspec

import (
	"bytes"
	"context"
	"net/http"
	"net/url"

	"github.com/adamluzsi/testcase"
)

var (
	Handler = testcase.Var{Name: `httpspec:Handler`}
	Context = testcase.Var{Name: `httpspec:Context`, Init: func(t *testcase.T) interface{} {
		return context.Background()
	}}
	Method = testcase.Var{Name: `httpspec:Method`, Init: func(t *testcase.T) interface{} {
		return http.MethodGet
	}}
	Path = testcase.Var{Name: `httpspec:Path`, Init: func(t *testcase.T) interface{} {
		return `/`
	}}
	Body = testcase.Var{Name: `httpspec:Body`, Init: func(t *testcase.T) interface{} {
		return &bytes.Buffer{}
	}}
	Query = testcase.Var{Name: `httpspec:QueryGet`, Init: func(t *testcase.T) interface{} {
		return url.Values{}
	}}
	Header = testcase.Var{Name: `httpspec:HeaderGet`, Init: func(t *testcase.T) interface{} {
		return http.Header{}
	}}
)

const (
	letVarPrefix   = `httpspec:`
	ContextVarName = letVarPrefix + `context`
)

// ContextGet allow to retrieve the current test scope's request context.
func ContextGet(t *testcase.T) context.Context {
	return Context.Get(t).(context.Context)
}

// QueryGet allows you to retrieve the current test scope's http PathGet query that will be used for SubjectGet.
// In a Before Block you can access the query and then specify the values in it.
func QueryGet(t *testcase.T) url.Values {
	return Query.Get(t).(url.Values)
}

// HeaderGet allows you to set the current test scope's http PathGet for SubjectGet.
func HeaderGet(t *testcase.T) http.Header {
	return Header.Get(t).(http.Header)
}


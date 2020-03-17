package httpspec_test

import (
	"net/http"
	"testing"
)

var testingT *testing.T

type MyHandler struct{}

func (m MyHandler) ServeHTTP(http.ResponseWriter, *http.Request) {}

type ListResponse struct {
	Resources []string
}

type ShowResponse struct {
	Resources []string
}

type CreateResponse struct {
	ID string
}

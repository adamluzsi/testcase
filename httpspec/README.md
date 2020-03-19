<!-- START doctoc generated TOC please keep comment here to allow auto update -->
<!-- DON'T EDIT THIS SECTION, INSTEAD RE-RUN doctoc TO UPDATE -->
**Table of Contents**  *generated with [DocToc](https://github.com/thlorenz/doctoc)*

- [httpspec](#httpspec)
  - [Documentation](#documentation)
  - [Usage](#usage)

<!-- END doctoc generated TOC please keep comment here to allow auto update -->

# httpspec

httpspec allow you to create HTTP API specifications with ease.

## [Documentation](https://godoc.org/github.com/adamluzsi/testcase/httpspec)

The documentation maintained in [GoDoc](https://godoc.org/github.com/adamluzsi/testcase/httpspec), including the [examples](https://godoc.org/github.com/adamluzsi/testcase/httpspec#pkg-examples).

## Usage

```go
package mypkg_test

import (
	"encoding/json"
	"net/http"
	"testing"

	"my/pkg/path/mypkg"

	"github.com/stretchr/testify/require"

	"github.com/adamluzsi/testcase"
	. "github.com/adamluzsi/testcase/httpspec"
)

func TestMyHandlerCreate(t *testing.T) {
	s := testcase.NewSpec(t)

	GivenThisIsAJSONAPI(s)

	// Arrange
	LetHandler(s, func(t *testcase.T) http.Handler { return mypkg.MyHandler{} })
	LetMethodValue(s, http.MethodPost)
	LetPathValue(s, `/`)
	LetBody(s, func(t *testcase.T) interface{} {
		// this will end up as {"foo":"bar"} in the request body
		return map[string]string{"foo": "bar"}
	})

	s.Then(`it will...`, func(t *testcase.T) {
		rr := ServeHTTP(t) // Act
		require.Equal(t, http.StatusOK, rr.Code)
		var resp mypkg.CreateResponse
		require.Nil(t, json.Unmarshal(rr.Body.Bytes(), &resp))
		// more assertion
	})
}
```

```go
package mypkg_test

import (
	"encoding/json"
	"net/http"
	"testing"

	"my/pkg/path/mypkg"

	"github.com/stretchr/testify/require"

	"github.com/adamluzsi/testcase"
	. "github.com/adamluzsi/testcase/httpspec"
)

func TestMyHandler(t *testing.T) {
	s := testcase.NewSpec(t)

	GivenThisIsAJSONAPI(s)

	LetHandler(s, func(t *testcase.T) http.Handler { return mypkg.MyHandler{} })

	s.Describe(`POST / - create X`, func(s *testcase.Spec) {
		LetMethodValue(s, http.MethodPost)
		LetPathValue(s, `/`)

		LetBody(s, func(t *testcase.T) interface{} {
			// this will end up as {"foo":"bar"} in the request body
			return map[string]string{"foo": "bar"}
		})

		var onSuccess = func(t *testcase.T) mypkg.CreateResponse {
			rr := ServeHTTP(t)
			require.Equal(t, http.StatusOK, rr.Code)
			var resp mypkg.CreateResponse
			require.Nil(t, json.Unmarshal(rr.Body.Bytes(), &resp))
			return resp
		}

		s.Then(`it will create a new resource`, func(t *testcase.T) {
			createResponse := onSuccess(t)
			// assert
			_ = createResponse
		})
	})
}
```
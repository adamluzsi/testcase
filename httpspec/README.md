<!-- START doctoc generated TOC please keep comment here to allow auto update -->
<!-- DON'T EDIT THIS SECTION, INSTEAD RE-RUN doctoc TO UPDATE -->

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
package mypkg

func TestMyHandlerCreate(t *testing.T) {
	s := testcase.NewSpec(t)

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
		assert.Must(t).Equal( http.StatusOK, rr.Code)
		var resp CreateResponse
		assert.Must(t).Nil( json.Unmarshal(rr.Body.Bytes(), &resp))
		// ...
	})
}
```

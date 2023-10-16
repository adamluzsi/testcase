package httpspec

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"

	"go.llib.dev/testcase"
)

var Body = testcase.Var[any]{ID: `httpspec:Body`, Init: func(t *testcase.T) any {
	return &bytes.Buffer{}
}}

func asIOReader(t *testcase.T, header http.Header, body any) (bodyValue io.ReadCloser) {
	defer func() {
		if !isDebugEnabled(t) {
			return
		}

		var buf bytes.Buffer
		_, err := io.Copy(&buf, bodyValue)
		if err != nil {
			t.Fatalf(`httpspec body debug print encountered an error: %v`, err.Error())
		}

		t.Log(`body:`)
		t.Log(buf.String())

		bodyValue = io.NopCloser(bytes.NewReader(buf.Bytes()))
	}()
	if body == nil {
		body = bytes.NewReader([]byte{})
	}
	if r, ok := body.(io.ReadCloser); ok {
		return r
	}
	if r, ok := body.(io.Reader); ok {
		return io.NopCloser(r)
	}

	var buf bytes.Buffer
	switch header.Get(`Content-Type`) {
	case `application/json`:
		if err := json.NewEncoder(&buf).Encode(body); err != nil {
			t.Fatalf(`httpspec request body creation encountered: %v`, err.Error())
		}

	case `application/x-www-form-urlencoded`:
		_, _ = fmt.Fprint(&buf, toURLValues(body).Encode())

	default:
		header.Set("Content-Type", "text/plain; charset=UTF-8")
		_, _ = fmt.Fprintf(&buf, "%v", body)

	}

	header.Add("Content-Length", strconv.Itoa(buf.Len()))

	return io.NopCloser(&buf)
}

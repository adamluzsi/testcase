package httpspec

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"strconv"

	"github.com/adamluzsi/testcase"
)

var Body = testcase.Var[any]{ID: `httpspec:Body`, Init: func(t *testcase.T) any {
	return &bytes.Buffer{}
}}

func bodyAsIOReader(t *testcase.T) (bodyValue io.Reader) {
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

		bodyValue = bytes.NewReader(buf.Bytes())
	}()

	if r, ok := Body.Get(t).(io.Reader); ok {
		return r
	}
	var buf bytes.Buffer
	switch Header.Get(t).Get(`Content-Type`) {
	case `application/json`:
		if err := json.NewEncoder(&buf).Encode(Body.Get(t)); err != nil {
			t.Fatalf(`httpspec request body creation encountered: %v`, err.Error())
		}

	case `application/x-www-form-urlencoded`:
		_, _ = fmt.Fprint(&buf, toURLValues(Body.Get(t)).Encode())
	}

	Header.Get(t).Add("Content-Length", strconv.Itoa(buf.Len()))

	return &buf
}

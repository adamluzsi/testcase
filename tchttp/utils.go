package tchttp

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"reflect"
	"strconv"

	"go.llib.dev/testcase"
)

func toURLValues(i interface{}) url.Values {
	if data, ok := i.(url.Values); ok {
		return data
	}

	rv := reflect.ValueOf(i)
	data := url.Values{}

	switch rv.Kind() {
	case reflect.Struct:
		rt := reflect.TypeOf(i)
		for i := 0; i < rv.NumField(); i++ {
			field := rv.Field(i)
			sf := rt.Field(i)
			var key string
			if nameInTag, ok := sf.Tag.Lookup(`form`); ok {
				key = nameInTag
			} else {
				key = sf.Name
			}
			data.Add(key, fmt.Sprint(field.Interface()))
		}

	case reflect.Map:
		for _, key := range rv.MapKeys() {
			mapValue := rv.MapIndex(key)
			switch mapValue.Kind() {
			case reflect.Slice:
				for i := 0; i < mapValue.Len(); i++ {
					data.Add(fmt.Sprint(key), fmt.Sprint(mapValue.Index(i).Interface()))
				}

			default:
				data.Add(fmt.Sprint(key), fmt.Sprint(mapValue.Interface()))
			}
		}

	case reflect.Ptr:
		for k, vs := range toURLValues(rv.Elem().Interface()) {
			for _, v := range vs {
				data.Add(k, v)
			}
		}

	}

	return data
}

func asIOReader(t *testcase.T, header http.Header, body any) (bodyValue io.ReadCloser) {
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

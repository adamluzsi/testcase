package httpspec

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"reflect"
	"strconv"

	"github.com/adamluzsi/testcase"
)

// ServeHTTP will make a request to the spec context
// it requires the following spec variables
//	* method -> http method <string>
//	* path -> http path <string>
//	* query -> http query string <url.Values>
//	* body -> http payload <io.Reader|io.ReadCloser>
//
func ServeHTTP(t *testcase.T) *httptest.ResponseRecorder {
	w := httptest.NewRecorder()
	target, _ := url.Parse(path(t))
	target.RawQuery = Query(t).Encode()
	if Debug {
		t.Log(`method:`, method(t))
		t.Log(`path`, target.String())
	}
	r := httptest.NewRequest(method(t), target.String(), bodyToIOReader(t))
	r = r.WithContext(ctx(t))
	r.Header = Header(t)
	handler(t).ServeHTTP(w, r)
	return w
}

func setup(s *testcase.Spec) {
	LetContext(s, func(t *testcase.T) context.Context { return context.Background() })
	LetMethod(s, func(t *testcase.T) string { return http.MethodGet })
	LetPath(s, func(t *testcase.T) string { return `/` })
	letQuery(s, func(t *testcase.T) url.Values { return url.Values{} })
	letHeader(s, func(t *testcase.T) http.Header { return http.Header{} })
	LetBody(s, func(t *testcase.T) interface{} { return &bytes.Buffer{} })
}

func bodyToIOReader(t *testcase.T) (bodyValue io.Reader) {
	defer func() {
		if !Debug {
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

	if r, ok := body(t).(io.Reader); ok {
		return r
	}
	var buf bytes.Buffer
	switch Header(t).Get(`Content-Type`) {
	case `application/json`:
		if err := json.NewEncoder(&buf).Encode(body(t)); err != nil {
			t.Fatalf(`httpspec request body creation encountered: %v`, err.Error())
		}

	case `application/x-www-form-urlencoded`:
		_, _ = fmt.Fprint(&buf, toURLValues(body(t)).Encode())
	}

	Header(t).Add("Content-Length", strconv.Itoa(buf.Len()))

	return &buf
}

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

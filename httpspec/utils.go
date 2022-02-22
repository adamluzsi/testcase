package httpspec

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/url"
	"reflect"
	"strconv"

	"github.com/adamluzsi/testcase"
)

func bodyToIOReader(t *testcase.T) (bodyValue io.Reader) {
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

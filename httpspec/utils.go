package httpspec

import (
	"fmt"
	"net/url"
	"reflect"
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

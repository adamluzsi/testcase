package reflects

import "reflect"

func IsNil(v any) bool {
	if v == nil {
		return true
	}
	defer func() { _ = recover() }()
	return reflect.ValueOf(v).IsNil()
}

func IsStruct(v any) bool {
	return reflect.ValueOf(v).Kind() == reflect.Struct
}

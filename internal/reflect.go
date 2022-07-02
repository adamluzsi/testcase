package internal

import "reflect"

func IsNil(v any) bool {
	defer func() { _ = recover() }()
	return reflect.ValueOf(v).IsNil()
}

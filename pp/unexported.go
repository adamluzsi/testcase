package pp

import (
	"reflect"
	"strings"
	"unsafe"
)

// TODO: figure out why some moq generated mocks have inf recursion that can avoid the recursion guard.
func canAccess(rv reflect.Value) bool {
	kind := rv.Kind()
	if rv.CanInterface() ||
		rv.CanAddr() ||
		rv.CanUint() ||
		rv.CanInt() ||
		rv.CanFloat() ||
		rv.CanComplex() ||
		kind == reflect.String ||
		kind == reflect.Array ||
		kind == reflect.Slice ||
		kind == reflect.Map ||
		kind == reflect.Interface ||
		kind == reflect.Pointer ||
		kind == reflect.Chan {
		return true
	}
	if kind == reflect.Struct {
		var firstChar string
		for _, char := range rv.Type().Name() {
			firstChar = string(char)
			break
		}
		if firstChar != "" && strings.ToUpper(firstChar) == firstChar {
			return true
		}
	}
	return false
}

func makeAccessable(rv reflect.Value) (reflect.Value, bool) {
	if rv.CanInterface() {
		return rv, true
	}
	if rv.CanAddr() {
		uv := reflect.NewAt(rv.Type(), unsafe.Pointer(rv.UnsafeAddr())).Elem()
		if uv.CanInterface() {
			return uv, true
		}
	}
	if rv.CanUint() {
		return reflect.ValueOf(rv.Uint()).Convert(rv.Type()), true
	}
	if rv.CanInt() {
		return reflect.ValueOf(rv.Int()).Convert(rv.Type()), true
	}
	if rv.CanFloat() {
		return reflect.ValueOf(rv.Float()).Convert(rv.Type()), true
	}
	if rv.CanComplex() {
		return reflect.ValueOf(rv.Complex()).Convert(rv.Type()), true
	}
	switch rv.Kind() {
	case reflect.String:
		return reflect.ValueOf(rv.String()).Convert(rv.Type()), true
	case reflect.Map:
		m := reflect.MakeMap(rv.Type())
		for _, key := range rv.MapKeys() {
			key, ok := makeAccessable(key)
			if !ok {
				continue
			}
			value, ok := makeAccessable(rv.MapIndex(key))
			if !ok {
				continue
			}
			m.SetMapIndex(key, value)
		}
		return m, true
	case reflect.Slice:
		slice := reflect.MakeSlice(rv.Type(), 0, rv.Len())
		for i, l := 0, rv.Len(); i < l; i++ {
			v, ok := makeAccessable(rv.Index(i))
			if !ok {
				continue
			}
			slice = reflect.Append(slice, v)
		}
		return slice, true
	}
	return rv, false
}

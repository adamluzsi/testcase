package reflects

import (
	"reflect"
)

func Accessible(rv reflect.Value) reflect.Value {
	if rv, ok := TryToMakeAccessible(rv); ok {
		return rv
	}
	return rv
}

func TryToMakeAccessible(rv reflect.Value) (reflect.Value, bool) {
	if rv.CanInterface() {
		return rv, true
	}
	if rv, ok := ToSettable(rv); ok {
		return rv, true
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
			key, ok := TryToMakeAccessible(key)
			if !ok {
				continue
			}
			value, ok := TryToMakeAccessible(rv.MapIndex(key))
			if !ok {
				continue
			}
			m.SetMapIndex(key, value)
		}
		return m, true
	case reflect.Slice:
		slice := reflect.MakeSlice(rv.Type(), 0, rv.Len())
		for i, l := 0, rv.Len(); i < l; i++ {
			v, ok := TryToMakeAccessible(rv.Index(i))
			if !ok {
				continue
			}
			slice = reflect.Append(slice, v)
		}
		return slice, true
	}
	return reflect.Value{}, false
}

func ToSettable(rv reflect.Value) (_ reflect.Value, ok bool) {
	if !rv.IsValid() {
		return reflect.Value{}, false
	}
	if rv.CanSet() {
		return rv, true
	}
	if rv.CanAddr() {
		if uv := reflect.NewAt(rv.Type(), rv.Addr().UnsafePointer()).Elem(); uv.CanInterface() {
			return uv, true
		}
	}
	return reflect.Value{}, false
}

package reflects

import "reflect"

func BaseTypeOf(v interface{}) reflect.Type {
	typ := reflect.TypeOf(v)
	for typ.Kind() == reflect.Ptr {
		typ = typ.Elem()
	}
	return typ
}

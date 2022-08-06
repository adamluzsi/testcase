package reflects

import (
	"reflect"
)

var acceptedConstKind = map[reflect.Kind]struct{}{
	reflect.String:     {},
	reflect.Bool:       {},
	reflect.Int:        {},
	reflect.Int8:       {},
	reflect.Int16:      {},
	reflect.Int32:      {},
	reflect.Int64:      {},
	reflect.Uint:       {},
	reflect.Uint8:      {},
	reflect.Uint16:     {},
	reflect.Uint32:     {},
	reflect.Uint64:     {},
	reflect.Float32:    {},
	reflect.Float64:    {},
	reflect.Complex64:  {},
	reflect.Complex128: {},
}

func IsMutable(v any) bool {
	if IsNil(v) {
		return false
	}
	return visitIsMutable(reflect.ValueOf(v))
}

func visitIsMutable(rv reflect.Value) bool {
	if rv.Kind() == reflect.Invalid {
		return false
	}
	if _, ok := acceptedConstKind[rv.Kind()]; ok {
		return false
	}
	if rv.Kind() == reflect.Struct {
		fieldNum := rv.NumField()
		for i, fNum := 0, fieldNum; i < fNum; i++ {
			name := rv.Type().Field(i).Name
			field := rv.FieldByName(name)
			if visitIsMutable(field) {
				return true
			}
		}
		return false
	}
	return true
}

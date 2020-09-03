package fixtures

import (
	"reflect"
	"time"
)

// New returns a populated entity for a given business data entity.
// This is primary and only used for testing.
// With fixtures, it become easy to create generic query objects that use test cases that does not specify the concrete Structure type.
func New(T interface{}) (pointer interface{}) {
	ptr := reflect.New(baseTypeOf(T))
	elem := ptr.Elem()

	for i := 0; i < elem.NumField(); i++ {
		fv := elem.Field(i)

		if fv.CanSet() {
			newValue := newValue(fv)

			if newValue.IsValid() {
				fv.Set(newValue)
			}
		}
	}

	return ptr.Interface()
}

func newValue(value reflect.Value) reflect.Value {
	switch value.Type().Kind() {

	case reflect.Bool:
		return reflect.ValueOf(Random.Bool())

	case reflect.String:
		return reflect.ValueOf(Random.String())

	case reflect.Int:
		return reflect.ValueOf(Random.Int())

	case reflect.Int8:
		return reflect.ValueOf(int8(Random.Int()))

	case reflect.Int16:
		return reflect.ValueOf(int16(Random.Int()))

	case reflect.Int32:
		return reflect.ValueOf(int32(Random.Int()))

	case reflect.Int64:
		switch value.Interface().(type) {
		case time.Duration:
			return reflect.ValueOf(time.Duration(Random.Int()))
		default:
			return reflect.ValueOf(int64(Random.Int()))
		}

	case reflect.Uint:
		return reflect.ValueOf(uint(Random.Int()))

	case reflect.Uint8:
		return reflect.ValueOf(uint8(Random.Int()))

	case reflect.Uint16:
		return reflect.ValueOf(uint16(Random.Int()))

	case reflect.Uint32:
		return reflect.ValueOf(uint32(Random.Int()))

	case reflect.Uint64:
		return reflect.ValueOf(uint64(Random.Int()))

	case reflect.Float32:
		return reflect.ValueOf(Random.Float32())

	case reflect.Float64:
		return reflect.ValueOf(Random.Float64())

	case reflect.Complex64:
		return reflect.ValueOf(complex64(42))

	case reflect.Complex128:
		return reflect.ValueOf(complex128(42.42))

	case reflect.Array:
		return reflect.New(value.Type()).Elem()

	case reflect.Slice:
		return reflect.MakeSlice(value.Type(), 0, 0)

	case reflect.Chan:
		return reflect.MakeChan(value.Type(), 0)

	case reflect.Map:
		return reflect.MakeMap(value.Type())

	case reflect.Ptr:
		return reflect.New(value.Type().Elem())

	case reflect.Uintptr:
		return reflect.ValueOf(uintptr(Random.Int()))

	case reflect.Struct:
		switch value.Interface().(type) {
		case time.Time:
			return reflect.ValueOf(Random.Time())
		default:
			return reflect.ValueOf(New(value.Interface())).Elem()
		}

	default:
		//reflect.UnsafePointer
		//reflect.Interface
		//reflect.Func
		//
		// returns nil to avoid unsafe edge cases
		return reflect.ValueOf(nil)
	}
}


func baseValueOf(i interface{}) reflect.Value {
	v := reflect.ValueOf(i)

	for v.Type().Kind() == reflect.Ptr {
		v = v.Elem()
	}

	return v
}

func baseTypeOf(i interface{}) reflect.Type {
	t := reflect.TypeOf(i)

	for t.Kind() == reflect.Ptr {
		t = t.Elem()
	}

	return t
}

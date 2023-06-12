package internal

import (
	"github.com/adamluzsi/testcase/internal/reflects"
	"github.com/adamluzsi/testcase/internal/teardown"
	"reflect"
)

func DeepEqual(v1, v2 any) (bool, error) {
	if v1 == nil || v2 == nil {
		return v1 == v2, nil
	}
	return reflectDeepEqual(
		&refMem{visited: make(map[uintptr]struct{})},
		reflect.ValueOf(v1), reflect.ValueOf(v2))
}

////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

func RegisterIsEqual(typ reflect.Type, rfn func(v1, v2 reflect.Value) (bool, error)) {
	isEqualFuncRegister[typ] = rfn
}

var isEqualFuncRegister = map[reflect.Type]func(v1, v2 reflect.Value) (bool, error){}

////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

func reflectDeepEqual(m *refMem, v1, v2 reflect.Value) (iseq bool, _ error) {
	if !m.TryVisit(v1, v2) {
		return true, nil // probably OK since we already visited it
	}
	if !v1.IsValid() || !v2.IsValid() {
		return v1.IsValid() == v2.IsValid(), nil
	}
	if v1.Type() != v2.Type() {
		return false, nil
	}
	if eq, err, ok := tryEqualityMethods(v1, v2); ok {
		return eq, err
	}

	switch v1.Kind() {
	case reflect.Struct:
		for i, n := 0, v1.NumField(); i < n; i++ {
			f1, ok := reflects.TryToMakeAccessible(v1.Field(i))
			if !ok {
				continue
			}
			f2, ok := reflects.TryToMakeAccessible(v2.Field(i))
			if !ok {
				continue
			}
			if eq, err := reflectDeepEqual(m, f1, f2); !eq {
				return eq, err
			}
		}
		return true, nil

	case reflect.Pointer:
		if v1.UnsafePointer() == v2.UnsafePointer() {
			return true, nil
		}
		return reflectDeepEqual(m, v1.Elem(), v2.Elem())

	case reflect.Array:
		// TODO: check if array with different length are considered as the same type
		for i := 0; i < v1.Len(); i++ {
			if eq, err := reflectDeepEqual(m, v1.Index(i), v2.Index(i)); !eq {
				return eq, err
			}
		}
		return true, nil

	case reflect.Slice:
		if v1.IsNil() != v2.IsNil() {
			return false, nil
		}
		if v1.Len() != v2.Len() {
			return false, nil
		}
		if v1.UnsafePointer() == v2.UnsafePointer() {
			return true, nil
		}
		// Special case for []byte, which is common.
		if v1.Type().Elem().Kind() == reflect.Uint8 {
			return string(v1.Bytes()) == string(v2.Bytes()), nil
		}
		for i := 0; i < v1.Len(); i++ {
			if eq, err := reflectDeepEqual(m, v1.Index(i), v2.Index(i)); !eq {
				return eq, err
			}
		}
		return true, nil

	case reflect.Interface:
		if v1.IsNil() || v2.IsNil() {
			return v1.IsNil() == v2.IsNil(), nil
		}
		return reflectDeepEqual(m, v1.Elem(), v2.Elem())

	case reflect.Map:
		if v1.IsNil() != v2.IsNil() {
			return false, nil
		}
		if v1.Len() != v2.Len() {
			return false, nil
		}
		if v1.UnsafePointer() == v2.UnsafePointer() {
			return true, nil
		}
		for _, k := range v1.MapKeys() {
			val1 := v1.MapIndex(k)
			val2 := v2.MapIndex(k)
			if !val1.IsValid() || !val2.IsValid() {
				return false, nil
			}
			if eq, err := reflectDeepEqual(m, val1, val2); !eq {
				return eq, err
			}
		}
		return true, nil

	case reflect.Func:
		if v1.IsNil() && v2.IsNil() {
			return true, nil
		}
		if v1.Pointer() == v2.Pointer() {
			return true, nil
		}
		return false, nil

	case reflect.Chan:
		if v1.IsNil() && v2.IsNil() {
			return true, nil
		}
		if v1.Cap() == 0 {
			return reflect.DeepEqual(v1.Interface(), v2.Interface()), nil
		}
		if v1.Cap() != v2.Cap() ||
			v1.Len() != v2.Len() {
			return false, nil
		}

		var (
			ln = v1.Len()
			td = &teardown.Teardown{}
		)
		defer td.Finish()
		for i := 0; i < ln; i++ {
			v1x, v1OK := v1.Recv()
			if v1OK {
				td.Defer(func() {
					v1.Send(v1x)
				})
			}
			v2x, v2OK := v1.Recv()
			if v2OK {
				td.Defer(func() {
					v2.Send(v2x)
				})
			}
			if v1OK != v2OK {
				return false, nil
			}
			if eq, err := reflectDeepEqual(m, v1x, v2x); !eq {
				return eq, err
			}
		}
		return true, nil

	default:
		return reflect.DeepEqual(
			reflects.Accessible(v1).Interface(),
			reflects.Accessible(v2).Interface()), nil
	}
}

////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

type refMem struct{ visited map[uintptr]struct{} }

func (i *refMem) TryVisit(v1, v2 reflect.Value) (ok bool) {
	return i.tryVisit(v1) || i.tryVisit(v2)
}

func (i *refMem) tryVisit(v reflect.Value) (ok bool) {
	if i.visited == nil {
		i.visited = make(map[uintptr]struct{})
	}
	key, ok := i.addr(v)
	if !ok {
		// for values that can't be tracked, we allow visiting
		// These are usually primitive types
		return true
	}
	if _, ok := i.visited[key]; ok {
		return false
	}
	i.visited[key] = struct{}{}
	return true
}

func (i *refMem) addr(v reflect.Value) (uintptr, bool) {
	switch v.Kind() {
	case reflect.Ptr, reflect.Slice, reflect.Map, reflect.Chan, reflect.Func, reflect.UnsafePointer:
		return v.Pointer(), true
	case reflect.Struct, reflect.Array:
		if v.CanAddr() {
			return v.Addr().Pointer(), true
		} else {
			return 0, false
		}
	default:
		// For basic types, use the address of the reflect.Value itself as the key.
		return reflect.ValueOf(&v).Pointer(), true
	}
}

func (i *refMem) canPointer(v reflect.Value) bool {
	switch v.Kind() {
	case reflect.Pointer, reflect.Chan, reflect.Map, reflect.UnsafePointer, reflect.Func, reflect.Slice:
		return true
	default:
		return false
	}
}

////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

func tryEqualityMethods(v1, v2 reflect.Value) (isEqual bool, _ error, ok bool) {
	defer func() { recover() }()
	if v1.Type() != v2.Type() {
		return false, nil, false
	}
	if eqfn, ok := isEqualFuncRegister[v1.Type()]; ok {
		isEq, err := eqfn(v1, v2)
		return isEq, err, true
	}
	if eq, err, ok := tryEquatable(v1, v2); ok {
		return eq, err, ok
	}
	if eq, ok := tryComparable(v1, v2); ok {
		return eq, nil, ok
	}
	return false, nil, false
}

func tryEquatable(v1, v2 reflect.Value) (bool, error, bool) {
	for _, methodName := range []string{"Equal", "IsEqual"} {
		if eq, err, ok := tryIsEqualMethod(methodName, v1, v2); ok {
			return eq, err, true
		}
		if eq, err, ok := tryIsEqualMethod(methodName, ptrOf(v1), v2); ok {
			return eq, err, true
		}
	}
	return false, nil, false
}

func ptrOf(v reflect.Value) reflect.Value {
	ptr := reflect.New(v.Type())
	ptr.Elem().Set(v)
	return ptr
}

var (
	errType  = reflect.TypeOf((*error)(nil)).Elem()
	boolType = reflect.TypeOf((*bool)(nil)).Elem()
	intType  = reflect.TypeOf((*int)(nil)).Elem()
)

func tryIsEqualMethod(methodName string, v1, v2 reflect.Value) (bool, error, bool) {
	method := v1.MethodByName(methodName)
	if method == (reflect.Value{}) {
		return false, nil, false
	}

	methodType := method.Type()

	if methodType.NumIn() != 1 {
		return false, nil, false
	}

	if methodType.In(0) != v2.Type() {
		return false, nil, false
	}

	if numOut := methodType.NumOut(); !(numOut == 1 || numOut == 2) {
		return false, nil, false
	}

	switch methodType.NumOut() {
	case 1:
		if methodType.Out(0) != boolType {
			return false, nil, false
		}
	case 2:
		if methodType.Out(0) != boolType {
			return false, nil, false
		}
		if !methodType.Out(1).Implements(errType) {
			return false, nil, false
		}
	default:
		return false, nil, false
	}

	result := method.Call([]reflect.Value{v2})

	switch methodType.NumOut() {
	case 1: // IsEqual(T) (bool)
		return result[0].Bool(), nil, true

	case 2: // IsEqual(T) (bool, error)
		eq := result[0].Bool()
		err, _ := result[1].Interface().(error)
		return eq, err, true

	default:
		return false, nil, false
	}
}

func tryComparable(v1, v2 reflect.Value) (bool, bool) {
	if eq, ok := tryCmp(v1, v2); ok {
		return eq, ok
	}
	if eq, ok := tryCmp(ptrOf(v1), v2); ok {
		return eq, ok
	}
	return false, false
}

func tryCmp(v1 reflect.Value, v2 reflect.Value) (bool, bool) {
	method := v1.MethodByName("Cmp")
	if method == (reflect.Value{}) {
		return false, false
	}
	methodType := method.Type()
	if methodType.NumIn() != 1 {
		return false, false
	}
	if methodType.In(0) != v2.Type() {
		return false, false
	}
	if methodType.NumOut() != 1 {
		return false, false
	}
	if methodType.Out(0) != intType {
		return false, false
	}
	result := method.Call([]reflect.Value{v2})
	return result[0].Int() == 0, true
}

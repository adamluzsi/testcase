package reflector

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

	if isEqual, err, ok := tryEqualityMethods(v1, v2); ok {
		return isEqual, err
	}

	switch v1.Kind() {
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

	case reflect.Pointer:
		if v1.UnsafePointer() == v2.UnsafePointer() {
			return true, nil
		}
		return reflectDeepEqual(m, v1.Elem(), v2.Elem())

	case reflect.Struct:
		for i, n := 0, v1.NumField(); i < n; i++ {
			var (
				f1 = reflects.Accessible(v1.Field(i))
				f2 = reflects.Accessible(v2.Field(i))
			)
			if eq, err := reflectDeepEqual(m, f1, f2); !eq {
				return eq, err
			}
		}
		return true, nil

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
		// Normal equality suffices
		return reflect.DeepEqual(v1.Interface(), v2.Interface()), nil
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

var methodNamesForIsEqual = []string{"IsEqual", "Equal"}

func tryEqualityMethods(v1, v2 reflect.Value) (isEqual bool, _ error, ok bool) {
	defer func() { recover() }()
	if v1.Type() != v2.Type() {
		return false, nil, false
	}
	if eqfn, ok := isEqualFuncRegister[v1.Type()]; ok {
		isEq, err := eqfn(v1, v2)
		return isEq, err, true
	}

	for _, methodName := range methodNamesForIsEqual {
		if eq, err, ok := tryIsEqualMethod(v1, v2, methodName); ok {
			return eq, err, ok
		}
	}
	return false, nil, false
}

var (
	errType  = reflect.TypeOf((*error)(nil)).Elem()
	boolType = reflect.TypeOf((*bool)(nil)).Elem()
)

func tryIsEqualMethod(v1, v2 reflect.Value, methodName string) (bool, error, bool) {
	method := v1.MethodByName(methodName)
	if method.IsZero() {
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

////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

var isEqualFuncRegister = map[reflect.Type]func(v1, v2 reflect.Value) (bool, error){}

func RegisterIsEqual(typ reflect.Type, rfn func(v1, v2 reflect.Value) (bool, error)) {
	isEqualFuncRegister[typ] = rfn
}

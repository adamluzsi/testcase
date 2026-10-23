package reflects

import (
	"fmt"
	"math"
	"reflect"
	"sync/atomic"
	"unsafe"
)

func DeepEqual(v1, v2 any) (bool, error) {
	if v1 == nil || v2 == nil {
		return v1 == v2, nil
	}
	return reflectDeepEqual(&refMem{},
		reflect.ValueOf(v1), reflect.ValueOf(v2))
}

////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

func RegisterIsEqual(typ reflect.Type, rfn func(v1, v2 reflect.Value) (bool, error)) {
	isEqualFuncRegister[typ] = rfn
}

var isEqualFuncRegister = map[reflect.Type]func(v1, v2 reflect.Value) (bool, error){}

////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

func reflectDeepEqual(m *refMem, v1, v2 reflect.Value) (iseq bool, _ error) {
	if !v1.IsValid() || !v2.IsValid() {
		return v1.IsValid() == v2.IsValid(), nil
	}
	if v1.Type() != v2.Type() {
		return false, nil
	}
	if m.Visited(v1, v2) {
		// This pair is already under comparison further up the call stack,
		// so it is a cycle, and that comparison will decide the outcome.
		return true, nil
	}
	if eq, err, ok := tryEqualityMethods(v1, v2); ok {
		return eq, err
	}

	switch v1.Kind() {
	case reflect.Struct:
		// The unexported fields of a non-addressable struct (e.g. one passed by value)
		// can't be made accessible, so we compare addressable copies instead.
		// This way every field is compared deeply, regardless of whether it is exported.
		v1, v2 = addressable(v1), addressable(v2)
		for i, n := 0, v1.NumField(); i < n; i++ {
			f1, ok1 := TryToMakeAccessible(v1.Field(i))
			f2, ok2 := TryToMakeAccessible(v2.Field(i))
			if !ok1 || !ok2 {
				return false, fmt.Errorf("unable to access %s.%s for comparison",
					v1.Type().String(), v1.Type().Field(i).Name)
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

	case reflect.Float32, reflect.Float64:
		// NaN != NaN under Go's == operator, which would make two values
		// that are both NaN compare as not-equal. For equality assertions
		// it is more useful to treat NaN as equal to NaN, so a computation
		// expected to yield NaN can be matched against NaN.
		f1, f2 := v1.Float(), v2.Float()
		if math.IsNaN(f1) && math.IsNaN(f2) {
			return true, nil
		}
		return f1 == f2, nil

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
		if v1.Pointer() == v2.Pointer() {
			return true, nil
		}
		if v1.Cap() == 0 {
			return reflect.DeepEqual(v1.Interface(), v2.Interface()), nil
		}
		if v1.Cap() != v2.Cap() ||
			v1.Len() != v2.Len() {
			return false, nil
		}
		// The buffered elements of a closed channel can't be put back once received,
		// so when either channel is closed, only their capacity and length must be the same.
		if !isOpenChan(v1) || !isOpenChan(v2) {
			return true, nil
		}
		// The buffered elements are fully drained before comparing them,
		// so they can be put back in their original order afterwards.
		vs1, vs2 := drainChan(v1), drainChan(v2)
		defer refillChan(v1, vs1)
		defer refillChan(v2, vs2)
		if len(vs1) != len(vs2) {
			return false, nil
		}
		for i := range vs1 {
			if eq, err := reflectDeepEqual(m, vs1[i], vs2[i]); !eq {
				return eq, err
			}
		}
		return true, nil

	default:
		return reflect.DeepEqual(
			Accessible(v1).Interface(),
			Accessible(v2).Interface()), nil
	}
}

////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

// refMem remembers which pairs of references are already under comparison,
// so cyclic data structures don't cause an infinite recursion.
//
// A visit is keyed by the address pair and the type together,
// because distinct values can share an address.
// For example, a pointer to a struct and a pointer to its first field.
type refMem struct{ visited map[visit]struct{} }

type visit struct {
	a1, a2 uintptr
	typ    reflect.Type
	// l1 and l2 tell apart slices that share a backing array but differ in length.
	l1, l2 int
}

// Visited reports whether the v1 and v2 pair was already visited,
// and marks it as visited.
func (m *refMem) Visited(v1, v2 reflect.Value) bool {
	switch v1.Kind() {
	case reflect.Pointer, reflect.Map, reflect.Slice, reflect.Chan:
		// only reference kinds can form a cycle
	default:
		return false
	}
	if v1.IsNil() || v2.IsNil() {
		return false
	}
	key := visit{a1: v1.Pointer(), a2: v2.Pointer(), typ: v1.Type()}
	if v1.Kind() == reflect.Slice {
		key.l1, key.l2 = v1.Len(), v2.Len()
	}
	if m.visited == nil {
		m.visited = make(map[visit]struct{})
	}
	if _, ok := m.visited[key]; ok {
		return true
	}
	m.visited[key] = struct{}{}
	return false
}

// addressable returns v, or an addressable copy of it when v is not addressable,
// so that its unexported fields can be made accessible.
func addressable(v reflect.Value) reflect.Value {
	if v.CanAddr() || !v.CanInterface() {
		return v
	}
	return ptrOf(v).Elem()
}

func drainChan(ch reflect.Value) []reflect.Value {
	var vs []reflect.Value
	for n := ch.Len(); len(vs) < n; {
		v, ok := ch.TryRecv()
		if !ok {
			break
		}
		vs = append(vs, v)
	}
	return vs
}

func refillChan(ch reflect.Value, vs []reflect.Value) {
	for _, v := range vs {
		ch.Send(v)
	}
}

// isOpenChan reports whether the ch channel is known to be open.
//
// Go has no API to tell whether a channel is closed without receiving from it,
// which would consume the buffered elements of a closed channel for good.
// Instead, the closed flag is read from the runtime's channel header.
// When the header layout can't be confirmed, every channel is treated as possibly closed.
func isOpenChan(ch reflect.Value) bool {
	return chanHeaderLayoutOK && !chanHeaderOf(ch).isClosed()
}

// chanHeader mirrors the leading fields of the runtime's channel representation (runtime.hchan).
type chanHeader struct {
	qcount   uint
	dataqsiz uint
	buf      unsafe.Pointer
	elemsize uint16
	closed   uint32
}

func chanHeaderOf(ch reflect.Value) *chanHeader {
	return (*chanHeader)(ch.UnsafePointer())
}

func (h *chanHeader) isClosed() bool {
	return atomic.LoadUint32(&h.closed) != 0
}

// chanHeaderLayoutOK tells whether chanHeader matches the runtime's channel layout,
// by checking it against a channel in a known state.
var chanHeaderLayoutOK = func() bool {
	ch := make(chan int, 2)
	ch <- 1
	h := chanHeaderOf(reflect.ValueOf(ch))
	if h.qcount != 1 || h.dataqsiz != 2 ||
		uintptr(h.elemsize) != unsafe.Sizeof(int(0)) || h.isClosed() {
		return false
	}
	close(ch)
	return h.isClosed()
}()

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
		method = v1.MethodByName("Compare") // fallback

		if method == (reflect.Value{}) {
			return false, false
		}
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

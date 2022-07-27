package pp

import (
	"bytes"
	"fmt"
	"io"
	"reflect"
	"sort"
	"strings"
	"sync"
	"time"
	"unsafe"
)

func Format(v any) string {
	return formatter{}.Format(v)
}

type formatter struct{}

func (f formatter) Format(v any) string {
	buf := &bytes.Buffer{}
	rv := reflect.ValueOf(v)
	(&visitor{}).Visit(buf, rv, 0)
	return buf.String()
}

type visitor struct {
	visitedInit sync.Once
	visited     map[reflect.Value]struct{}
}

func (vis *visitor) Visit(w io.Writer, rv reflect.Value, depth int) {
	td, ok := vis.recursionGuard(w, rv)
	if !ok {
		return
	}
	defer td()

	if rv.Kind() == reflect.Invalid {
		fmt.Fprint(w, "nil")
		return
	}

	if rv.CanInt() {
		fmt.Fprintf(w, "%#v", rv.Int())
		return
	}
	if rv.CanUint() {
		fmt.Fprintf(w, "%d", rv.Uint())
		return
	}
	if rv.CanFloat() {
		fmt.Fprintf(w, "%#v", rv.Float())
		return
	}

	switch rv.Kind() {
	case reflect.Array, reflect.Slice:
		if vis.tryStringer(w, rv, depth) {
			return
		}

		fmt.Fprintf(w, "%s{", rv.Type().String())
		vLen := rv.Len()
		for i := 0; i < vLen; i++ {
			vis.newLine(w, depth+1)
			vis.Visit(w, rv.Index(i), depth+1)
			fmt.Fprintf(w, ",")
		}
		if 0 < vLen {
			vis.newLine(w, depth)
		}
		fmt.Fprint(w, "}")

	case reflect.Map:
		fmt.Fprintf(w, "%s{", rv.Type().String())
		keys := rv.MapKeys()
		vis.sortMapKeys(keys)
		for _, key := range keys {
			vis.newLine(w, depth+1)
			vis.Visit(w, key, depth+1) // key
			fmt.Fprintf(w, ": ")
			vis.Visit(w, rv.MapIndex(key), depth+1) // value
			fmt.Fprintf(w, ",")
		}
		if 0 < len(keys) {
			vis.newLine(w, depth)
		}
		fmt.Fprint(w, "}")

	case reflect.Struct:
		switch rv.Type() {
		case reflect.TypeOf(time.Time{}):
			fmt.Fprintf(w, "%#v", rv.Interface())
		default:
			vis.visitGenericStructure(w, rv, depth)
		}
	case reflect.Interface:
		fmt.Fprintf(w, "(%s)(", rv.Type().String())
		vis.Visit(w, rv.Elem(), depth)
		fmt.Fprint(w, ")")

	case reflect.Pointer:
		if rv.IsNil() {
			vis.Visit(w, reflect.ValueOf(nil), depth)
			return
		}

		elem := rv.Elem()
		if vis.isRecursion(elem) {
			fmt.Fprintf(w, "(%s)(", rv.Type().String())
			fmt.Fprintf(w, "%#v", rv.Pointer())
			fmt.Fprint(w, ")")
			return
		}

		fmt.Fprintf(w, "&")
		vis.Visit(w, rv.Elem(), depth)

	case reflect.String:
		fmt.Fprintf(w, "%#v", rv.String())

	default:
		v, ok := vis.makeAccessable(rv)
		if !ok {
			fmt.Fprint(w, "/* inaccessible */")
			return
		}
		fmt.Fprintf(w, "%#v", v.Interface())
	}
}

func (vis *visitor) recursionGuard(w io.Writer, rv reflect.Value) (_td func(), _ok bool) {
	vis.visitedInit.Do(func() { vis.visited = make(map[reflect.Value]struct{}) })
	if !vis.isRecursion(rv) {
		vis.visited[rv] = struct{}{}
		return func() { delete(vis.visited, rv) }, true
	}
	if rv.CanAddr() {
		fmt.Fprintf(w, "%#v", rv.UnsafeAddr())
	} else if rv.CanInterface() {
		fmt.Fprintf(w, "%#v", rv.Interface())
	} else {
		fmt.Fprintf(w, "%v", rv)
	}
	return func() {}, false
}

func (vis *visitor) isRecursion(v reflect.Value) bool {
	if vis.visited == nil {
		return false
	}
	_, ok := vis.visited[v]
	return ok
}

func (vis *visitor) isEmpty(v reflect.Value) bool {
	if vis.isNil(v) {
		return true
	}
	rv := reflect.ValueOf(v)
	switch rv.Kind() {
	case reflect.Chan, reflect.Map, reflect.Slice:
		return rv.Len() == 0

	case reflect.Array:
		zero := reflect.New(rv.Type()).Elem().Interface()
		return reflect.DeepEqual(zero, v)

	case reflect.Ptr:
		if rv.IsNil() {
			return true
		}
		return vis.isEmpty(rv.Elem())

	default:
		return reflect.DeepEqual(reflect.Zero(rv.Type()).Interface(), v)
	}
}

func (vis *visitor) isNil(v reflect.Value) bool {
	defer func() { _ = recover() }()
	if v.CanInterface() && v.Interface() == nil {
		return true
	}
	return v.IsNil()
}

var fmtStringerType = reflect.TypeOf((*fmt.Stringer)(nil)).Elem()

func (vis visitor) tryStringer(w io.Writer, v reflect.Value, depth int) bool {
	if !v.Type().Implements(fmtStringerType) {
		return false
	}
	vis.Visit(w, v.MethodByName("String").Call([]reflect.Value{})[0], depth)
	return true
}

func (vis visitor) visitGenericStructure(w io.Writer, v reflect.Value, depth int) {
	fmt.Fprintf(w, "%s{", v.Type().String())
	fieldNum := v.NumField()
	for i, fNum := 0, fieldNum; i < fNum; i++ {
		name := v.Type().Field(i).Name
		field := v.FieldByName(name)
		// if reflect pkg change and int and other values no longer be accessible, then this can skip unexported fields
		//if !field.CanInterface() {
		//	continue
		//}
		vis.newLine(w, depth+1)
		fmt.Fprintf(w, "%s: ", name)
		vis.Visit(w, field, depth+1)
		fmt.Fprintf(w, ",")
	}
	if 0 < fieldNum {
		vis.newLine(w, depth)
	}
	fmt.Fprint(w, "}")
}

func (vis visitor) newLine(w io.Writer, depth int) {
	_, _ = w.Write([]byte("\n"))
	vis.indent(w, depth)
}

func (vis visitor) indent(w io.Writer, depth int) {
	const defaultIndent = "\t"
	_, _ = w.Write([]byte(strings.Repeat(defaultIndent, depth)))
}

func (vis visitor) sortMapKeys(keys []reflect.Value) {
	if 0 == len(keys) {
		return
	}
	kind := keys[0].Kind()
	sort.Slice(keys, func(i, j int) bool {
		switch kind {
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			return keys[i].Int() < keys[j].Int()
		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			return keys[i].Uint() < keys[j].Uint()
		case reflect.Float32, reflect.Float64:
			return keys[i].Float() < keys[j].Float()
		case reflect.String:
			return keys[i].String() < keys[j].String()
		default:
			return Format(keys[i]) < Format(keys[j])
		}
	})
}

func (vis visitor) makeAccessable(v reflect.Value) (reflect.Value, bool) {
	if v.CanInterface() {
		return v, true
	}
	if v.CanAddr() {
		uv := reflect.NewAt(v.Type(), unsafe.Pointer(v.UnsafeAddr())).Elem()
		if uv.CanInterface() {
			return uv, true
		}
	}
	return v, false
}

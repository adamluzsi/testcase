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

func (vis *visitor) Visit(w io.Writer, v reflect.Value, depth int) {
	if vis.isVisited(w, v) {
		return
	}
	if v.Kind() == reflect.Invalid {
		fmt.Fprint(w, "nil")
		return
	}

	if v.CanInt() {
		fmt.Fprintf(w, "%#v", v.Int())
		return
	}
	if v.CanUint() {
		fmt.Fprintf(w, "%d", v.Uint())
		return
	}
	if v.CanFloat() {
		fmt.Fprintf(w, "%#v", v.Float())
		return
	}

	switch v.Kind() {
	case reflect.Array, reflect.Slice:
		if vis.tryStringer(w, v, depth) {
			return
		}

		fmt.Fprintf(w, "%s{", v.Type().String())
		vLen := v.Len()
		for i := 0; i < vLen; i++ {
			vis.newLine(w, depth+1)
			vis.Visit(w, v.Index(i), depth+1)
			fmt.Fprintf(w, ",")
		}
		if 0 < vLen {
			vis.newLine(w, depth)
		}
		fmt.Fprint(w, "}")

	case reflect.Map:
		fmt.Fprintf(w, "%s{", v.Type().String())
		keys := v.MapKeys()
		vis.sortMapKeys(keys)
		for _, key := range keys {
			vis.newLine(w, depth+1)
			vis.Visit(w, key, depth+1) // key
			fmt.Fprintf(w, ": ")
			vis.Visit(w, v.MapIndex(key), depth+1) // value
			fmt.Fprintf(w, ",")
		}
		if 0 < len(keys) {
			vis.newLine(w, depth)
		}
		fmt.Fprint(w, "}")

	case reflect.Struct:
		switch v.Type() {
		case reflect.TypeOf(time.Time{}):
			fmt.Fprintf(w, "%#v", v.Interface())
		default:
			vis.visitGenericStructure(w, v, depth)
		}
	case reflect.Interface:
		fmt.Fprintf(w, "(%s)(", v.Type().String())
		vis.Visit(w, v.Elem(), depth)
		fmt.Fprint(w, ")")

	case reflect.Pointer:
		if v.IsNil() {
			vis.Visit(w, reflect.ValueOf(nil), depth)
			return
		}

		fmt.Fprintf(w, "&")
		vis.Visit(w, v.Elem(), depth)

	case reflect.String:
		fmt.Fprintf(w, "%#v", v.String())

	default:
		v, ok := vis.makeAccessable(v)
		if !ok {
			fmt.Fprint(w, "/* inaccessible */")
			return
		}
		fmt.Fprintf(w, "%#v", v.Interface())
	}
}

func (vis *visitor) isVisited(w io.Writer, v reflect.Value) bool {
	vis.visitedInit.Do(func() { vis.visited = make(map[reflect.Value]struct{}) })

	_, ok := vis.visited[v]
	if !ok {
		vis.visited[v] = struct{}{}
		return false
	}
	if v == reflect.ValueOf(struct{}{}) {
		return false
	}
	if vis.isEmpty(v) {
		return false
	}

	fmt.Fprint(w, "/* recursion */")
	return true
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

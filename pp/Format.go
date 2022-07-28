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
	"unicode/utf8"
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

func (v *visitor) Visit(w io.Writer, rv reflect.Value, depth int) {
	defer debugRecover()
	td, ok := v.recursionGuard(w, rv)
	if !ok {
		return
	}
	defer td()

	if rv.Kind() == reflect.Invalid {
		fmt.Fprint(w, "nil")
		return
	}

	rv, _ = makeAccessable(rv)

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
		if v.tryStringer(w, rv, depth) {
			return
		}
		if v.tryByteSlice(w, rv) {
			return
		}

		fmt.Fprintf(w, "%s{", v.getTypeName(rv))
		vLen := rv.Len()
		for i := 0; i < vLen; i++ {
			v.newLine(w, depth+1)
			v.Visit(w, rv.Index(i), depth+1)
			fmt.Fprintf(w, ",")
		}
		if 0 < vLen {
			v.newLine(w, depth)
		}
		fmt.Fprint(w, "}")

	case reflect.Map:
		fmt.Fprintf(w, "%s{", v.getTypeName(rv))
		keys := rv.MapKeys()
		v.sortMapKeys(keys)
		for _, key := range keys {
			v.newLine(w, depth+1)
			v.Visit(w, key, depth+1) // key
			fmt.Fprintf(w, ": ")
			v.Visit(w, rv.MapIndex(key), depth+1) // value
			fmt.Fprintf(w, ",")
		}
		if 0 < len(keys) {
			v.newLine(w, depth)
		}
		fmt.Fprint(w, "}")

	case reflect.Struct:
		switch rv.Type() {
		case reflect.TypeOf(time.Time{}):
			fmt.Fprintf(w, "%#v", rv.Interface())
		default:
			v.visitGenericStructure(w, rv, depth)
		}
	case reflect.Interface:
		fmt.Fprintf(w, "(%s)(", v.getTypeName(rv))
		v.Visit(w, rv.Elem(), depth)
		fmt.Fprint(w, ")")

	case reflect.Pointer:
		if rv.IsNil() {
			v.Visit(w, reflect.ValueOf(nil), depth)
			return
		}

		elem := rv.Elem()
		if v.isRecursion(elem) {
			fmt.Fprintf(w, "(%s)(", v.getTypeName(rv))
			fmt.Fprintf(w, "%#v", rv.Pointer())
			fmt.Fprint(w, ")")
			return
		}

		fmt.Fprintf(w, "&")
		v.Visit(w, rv.Elem(), depth)

	case reflect.Chan:
		fmt.Fprintf(w, "make(%s, %d)", rv.Type().String(), rv.Cap())

	case reflect.String:
		fmt.Fprintf(w, "%#v", rv.String())

	default:
		v, ok := makeAccessable(rv)
		if !ok {
			fmt.Fprint(w, "/* inaccessible */")
			return
		}
		fmt.Fprintf(w, "%#v", v.Interface())
	}
}

func (v *visitor) recursionGuard(w io.Writer, rv reflect.Value) (_td func(), _ok bool) {
	v.visitedInit.Do(func() { v.visited = make(map[reflect.Value]struct{}) })
	if !v.isRecursion(rv) {
		v.visited[rv] = struct{}{}
		return func() { delete(v.visited, rv) }, true
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

func (v *visitor) isRecursion(rv reflect.Value) bool {
	if v.visited == nil {
		return false
	}
	_, ok := v.visited[rv]
	return ok
}

func (v *visitor) isEmpty(rv reflect.Value) bool {
	if v.isNil(rv) {
		return true
	}
	switch rv.Kind() {
	case reflect.Chan, reflect.Map, reflect.Slice:
		return rv.Len() == 0

	case reflect.Array:
		zero := reflect.New(rv.Type()).Elem().Interface()
		return reflect.DeepEqual(zero, rv)

	case reflect.Ptr:
		if rv.IsNil() {
			return true
		}
		return v.isEmpty(rv.Elem())

	default:
		return reflect.DeepEqual(reflect.Zero(rv.Type()).Interface(), rv)
	}
}

func (v *visitor) isNil(rv reflect.Value) bool {
	defer func() { _ = recover() }()
	if rv.CanInterface() && rv.Interface() == nil {
		return true
	}
	return rv.IsNil()
}

var fmtStringerType = reflect.TypeOf((*fmt.Stringer)(nil)).Elem()

func (v visitor) tryStringer(w io.Writer, rv reflect.Value, depth int) bool {
	if !rv.Type().Implements(fmtStringerType) {
		return false
	}

	fmt.Fprintf(w, "/* %s */ ", rv.Type().String())
	v.Visit(w, rv.MethodByName("String").Call([]reflect.Value{})[0], depth)
	return true
}

func (v visitor) visitGenericStructure(w io.Writer, rv reflect.Value, depth int) {
	fmt.Fprintf(w, "%s{", rv.Type().String())
	fieldNum := rv.NumField()
	for i, fNum := 0, fieldNum; i < fNum; i++ {
		name := rv.Type().Field(i).Name
		field := rv.FieldByName(name)
		v.newLine(w, depth+1)
		fmt.Fprintf(w, "%s: ", name)
		v.Visit(w, field, depth+1)
		fmt.Fprintf(w, ",")
	}
	if 0 < fieldNum {
		v.newLine(w, depth)
	}
	fmt.Fprint(w, "}")
}

func (v visitor) newLine(w io.Writer, depth int) {
	_, _ = w.Write([]byte("\n"))
	v.indent(w, depth)
}

func (v visitor) indent(w io.Writer, depth int) {
	const defaultIndent = "\t"
	_, _ = w.Write([]byte(strings.Repeat(defaultIndent, depth)))
}

func (v visitor) sortMapKeys(keys []reflect.Value) {
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

var typeByteSlice = reflect.TypeOf([]byte{})

func (v *visitor) tryByteSlice(w io.Writer, rv reflect.Value) bool {
	if !rv.Type().ConvertibleTo(typeByteSlice) {
		return false
	}

	var data = rv.Convert(typeByteSlice).Bytes()
	if !utf8.Valid(data) {
		return false
	}

	var (
		typeName  = v.getTypeName(rv)
		quoteChar = "`"
		content   = string(data)
	)
	switch {
	case !strings.Contains(content, `"`):
		quoteChar = `"`
	case !strings.Contains(content, "`"):
		quoteChar = "`"
	default:
		quoteChar = "`"
		content = strings.ReplaceAll(content, "`", "`+\"`\"+`")
	}

	fmt.Fprintf(w, "%s(%s%s%s)", typeName, quoteChar, content, quoteChar)
	return true
}

func (v *visitor) getTypeName(rv reflect.Value) string {
	var typeName = rv.Type().String()
	if rv.Type() == typeByteSlice {
		typeName = "[]byte"
	}
	return typeName
}

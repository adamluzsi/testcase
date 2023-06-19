package pp

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/adamluzsi/testcase/internal/reflects"
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

var (
	typeByteSlice    = reflect.TypeOf((*[]byte)(nil)).Elem()
	typeTimeDuration = reflect.TypeOf((*time.Duration)(nil)).Elem()
	typeTimeTime     = reflect.TypeOf((*time.Time)(nil)).Elem()
)

type formatter struct{}

func (f formatter) Format(v any) string {
	buf := &bytes.Buffer{}
	rv := reflect.ValueOf(v)
	vis := &visitor{}
	vis.Visit(buf, rv, 0)
	if vis.isStackoverflow() {
		return fmt.Sprintf("%#v", v)
	}
	return buf.String()
}

type visitor struct {
	visitedInit sync.Once
	visited     map[reflect.Value]struct{}
	stack       int
}

func (v *visitor) Visit(w io.Writer, rv reflect.Value, depth int) {
	defer debugRecover()
	td, ok := v.recursionGuard(w, rv)
	if !ok {
		return
	}
	defer td()

	if rv.Kind() == reflect.Invalid {
		_, _ = fmt.Fprint(w, "nil")
		return
	}

	rv = reflects.Accessible(rv)

	if rv.Type() == typeTimeDuration {
		d := time.Duration(rv.Int())
		_, _ = fmt.Fprintf(w, "/* %s */ %#v", d.String(), d)
		return
	}

	if rv.Type() == typeTimeTime {
		_, _ = fmt.Fprintf(w, "%#v", rv.Interface())
		return
	}

	if v.tryStringer(w, rv, depth) {
		return
	}

	if rv.CanInt() {
		_, _ = fmt.Fprintf(w, "%#v", rv.Int())
		return
	}

	if rv.CanUint() {
		_, _ = fmt.Fprintf(w, "%d", rv.Uint())
		return
	}

	if rv.CanFloat() {
		_, _ = fmt.Fprintf(w, "%#v", rv.Float())
		return
	}

	switch rv.Kind() {
	case reflect.Array, reflect.Slice:
		if v.tryNilSlice(w, rv) {
			return
		}
		if v.tryByteSlice(w, rv, depth) {
			return
		}

		_, _ = fmt.Fprintf(w, "%s{", v.getTypeName(rv))
		vLen := rv.Len()
		for i := 0; i < vLen; i++ {
			v.newLine(w, depth+1)
			v.Visit(w, rv.Index(i), depth+1)
			_, _ = fmt.Fprintf(w, ",")
		}
		if 0 < vLen {
			v.newLine(w, depth)
		}
		_, _ = fmt.Fprint(w, "}")

	case reflect.Map:
		_, _ = fmt.Fprintf(w, "%s{", v.getTypeName(rv))
		keys := rv.MapKeys()
		v.sortMapKeys(keys)
		for _, key := range keys {
			v.newLine(w, depth+1)
			v.Visit(w, key, depth+1) // key
			_, _ = fmt.Fprintf(w, ": ")
			v.Visit(w, rv.MapIndex(key), depth+1) // value
			_, _ = fmt.Fprintf(w, ",")
		}
		if 0 < len(keys) {
			v.newLine(w, depth)
		}
		_, _ = fmt.Fprint(w, "}")

	case reflect.Struct:
		v.visitStructure(w, rv, depth)

	case reflect.Interface:
		_, _ = fmt.Fprintf(w, "(%s)(", v.getTypeName(rv))
		v.Visit(w, rv.Elem(), depth)
		_, _ = fmt.Fprint(w, ")")

	case reflect.Pointer:
		if rv.IsNil() {
			v.Visit(w, reflect.ValueOf(nil), depth)
			return
		}

		elem := rv.Elem()
		if v.isRecursion(elem) {
			_, _ = fmt.Fprintf(w, "(%s)(", v.getTypeName(rv))
			_, _ = fmt.Fprintf(w, "%#v", rv.Pointer())
			_, _ = fmt.Fprint(w, ")")
			return
		}

		_, _ = fmt.Fprintf(w, "&")
		v.Visit(w, rv.Elem(), depth)

	case reflect.Chan:
		_, _ = fmt.Fprintf(w, "make(%s, %d)", rv.Type().String(), rv.Cap())

	case reflect.String:
		_, _ = fmt.Fprintf(w, "%#v", rv.String())

	default:
		v, ok := reflects.TryToMakeAccessible(rv)
		if !ok {
			_, _ = fmt.Fprint(w, "/* inaccessible */")
			return
		}
		_, _ = fmt.Fprintf(w, "%#v", v.Interface())
	}
}

func (v *visitor) recursionGuard(w io.Writer, rv reflect.Value) (_td func(), _ok bool) {
	v.stack++
	if v.isStackoverflow() {
		return
	}
	v.visitedInit.Do(func() { v.visited = make(map[reflect.Value]struct{}) })
	if !v.isRecursion(rv) {
		v.visited[rv] = struct{}{}
		return func() { delete(v.visited, rv) }, true
	}
	if rv.CanAddr() {
		_, _ = fmt.Fprintf(w, "%#v", rv.UnsafeAddr())
	} else if rv.CanInterface() {
		_, _ = fmt.Fprintf(w, "%#v", rv.Interface())
	} else {
		_, _ = fmt.Fprintf(w, "%v", rv)
	}
	return func() {}, false
}

func (v *visitor) isStackoverflow() bool {
	return 256 < v.stack
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

func (v *visitor) tryStringer(w io.Writer, rv reflect.Value, depth int) bool {
	if !rv.Type().Implements(fmtStringerType) {
		return false
	}

	_, _ = fmt.Fprintf(w, "/* %s */ ", rv.Type().String())
	v.Visit(w, rv.MethodByName("String").Call([]reflect.Value{})[0], depth)
	return true
}

func (v *visitor) visitStructure(w io.Writer, rv reflect.Value, depth int) {
	_, _ = fmt.Fprintf(w, "%s{", rv.Type().String())
	fieldNum := rv.NumField()
	for i, fNum := 0, fieldNum; i < fNum; i++ {
		name := rv.Type().Field(i).Name
		field := rv.FieldByName(name)

		v.newLine(w, depth+1)
		_, _ = fmt.Fprintf(w, "%s: ", name)
		v.Visit(w, field, depth+1)
		_, _ = fmt.Fprintf(w, ",")
	}
	if 0 < fieldNum {
		v.newLine(w, depth)
	}
	_, _ = fmt.Fprint(w, "}")
}

func (v *visitor) newLine(w io.Writer, depth int) {
	_, _ = w.Write([]byte("\n"))
	_, _ = w.Write([]byte(v.indent(depth)))
}

func (v *visitor) indent(depth int) string {
	const defaultIndent = "\t"
	return strings.Repeat(defaultIndent, depth)
}

func (v *visitor) sortMapKeys(keys []reflect.Value) {
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

func (v *visitor) tryByteSlice(w io.Writer, rv reflect.Value, depth int) bool {
	if !rv.Type().ConvertibleTo(typeByteSlice) {
		return false
	}

	var data = rv.Convert(typeByteSlice).Bytes()
	if !utf8.Valid(data) {
		return false
	}

	if json.Valid(data) {
		var buf bytes.Buffer
		if err := json.Indent(&buf, data, v.indent(depth), "\t"); err == nil {
			data = buf.Bytes()
		}
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

	_, _ = fmt.Fprintf(w, "%s(%s%s%s)", typeName, quoteChar, content, quoteChar)
	return true
}

func (v *visitor) getTypeName(rv reflect.Value) string {
	var typeName = rv.Type().String()
	if rv.Type() == typeByteSlice {
		typeName = "[]byte"
	}
	return typeName
}

func (v *visitor) tryNilSlice(w io.Writer, rv reflect.Value) bool {
	if rv.Kind() != reflect.Slice || !rv.IsNil() {
		return false
	}
	_, _ = fmt.Fprintf(w, "(%s)(nil)", rv.Type().String())
	return true
}

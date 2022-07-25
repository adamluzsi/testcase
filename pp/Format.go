package pp

import (
	"bytes"
	"fmt"
	"io"
	"reflect"
	"sort"
	"strings"
	"unsafe"
)

func Format(v any) string {
	return formatter{}.Format(v)
}

type formatter struct{}

func (f formatter) Format(v any) string {
	buf := &bytes.Buffer{}
	rv := reflect.ValueOf(v)
	f.visit(buf, rv, 0)
	return buf.String()
}

func (f formatter) visit(w io.Writer, v reflect.Value, depth int) {
	switch v.Kind() {
	case reflect.Array, reflect.Slice:
		fmt.Fprintf(w, "%s{", v.Type().String())
		vLen := v.Len()
		for i := 0; i < vLen; i++ {
			f.newLine(w, depth+1)
			f.visit(w, v.Index(i), depth+1)
			fmt.Fprintf(w, ",")
		}
		if 0 < vLen {
			f.newLine(w, depth)
		}
		fmt.Fprint(w, "}")

	case reflect.Map:
		fmt.Fprintf(w, "%s{", v.Type().String())
		keys := v.MapKeys()
		f.sortMapKeys(keys)
		for _, key := range keys {
			f.newLine(w, depth+1)
			f.visit(w, key, depth+1) // key
			fmt.Fprintf(w, ": ")
			f.visit(w, v.MapIndex(key), depth+1) // value
			fmt.Fprintf(w, ",")
		}
		if 0 < len(keys) {
			f.newLine(w, depth)
		}
		fmt.Fprint(w, "}")

	case reflect.Struct:
		// hack, cleanup this with recursion handling
		_ = fmt.Sprintf("%#v", v.Interface())

		fmt.Fprintf(w, "%s{", v.Type().String())
		fieldNum := v.NumField()
		for i, fNum := 0, fieldNum; i < fNum; i++ {
			name := v.Type().Field(i).Name
			field := v.FieldByName(name)
			// if reflect pkg change and int and other values no longer be accessible, then this can skip unexported fields
			//if !field.CanInterface() {
			//	continue
			//}
			f.newLine(w, depth+1)
			fmt.Fprintf(w, "%s: ", name)
			f.visit(w, field, depth+1)
			fmt.Fprintf(w, ",")
		}
		if 0 < fieldNum {
			f.newLine(w, depth)
		}
		fmt.Fprint(w, "}")

	case reflect.Interface:
		fmt.Fprintf(w, "(%s)(", v.Type().String())
		f.visit(w, v.Elem(), depth)
		fmt.Fprint(w, ")")

	case reflect.Pointer:
		fmt.Fprintf(w, "&")
		f.visit(w, v.Elem(), depth)

	case reflect.Invalid:
		fmt.Fprint(w, "nil")

	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		fmt.Fprintf(w, "%#v", v.Int())

	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		fmt.Fprintf(w, "%#v", v.Uint())

	case reflect.Float32, reflect.Float64:
		fmt.Fprintf(w, "%#v", v.Float())

	case reflect.String:
		fmt.Fprintf(w, "%#v", v.String())

	default:
		if v.CanInterface() {
			fmt.Fprintf(w, "%#v", v.Interface())
		} else {
			fmt.Fprint(w, "<unaccessible>")
		}
	}
}

func (f formatter) newLine(w io.Writer, depth int) {
	_, _ = w.Write([]byte("\n"))
	f.indent(w, depth)
}

func (f formatter) indent(w io.Writer, depth int) {
	const defaultIndent = "\t"
	_, _ = w.Write([]byte(strings.Repeat(defaultIndent, depth)))
}

func (f formatter) sortMapKeys(keys []reflect.Value) {
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
		}
		return false
	})
}

func (f formatter) getUnexportedValue(rf reflect.Value) reflect.Value {
	return reflect.NewAt(rf.Type(), unsafe.Pointer(rf.UnsafeAddr())).Elem()
}

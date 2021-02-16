package testcase

import (
	"fmt"
	"reflect"
)

func fullyQualifiedName(e interface{}) string {
	typ := baseTypeOf(e)

	if typ.PkgPath() == "" {
		return fmt.Sprintf("%s", typ.Name())
	}

	return fmt.Sprintf("%q.%s", typ.PkgPath(), typ.Name())
}

func baseTypeOf(i interface{}) reflect.Type {
	t := reflect.TypeOf(i)

	for t.Kind() == reflect.Ptr {
		t = t.Elem()
	}

	return t
}

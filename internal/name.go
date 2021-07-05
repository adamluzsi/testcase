package internal

import (
	"fmt"
	"path/filepath"
	"reflect"
)

func SymbolicName(T interface{}) string {
	t := reflect.TypeOf(T)
	for t.Kind() == reflect.Ptr {
		t = t.Elem()
	}

	if t.PkgPath() == "" {
		return fmt.Sprintf("%s", t.Name())
	}

	return fmt.Sprintf("%s.%s", filepath.Base(t.PkgPath()), t.Name())
}

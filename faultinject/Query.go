package faultinject

import (
	"github.com/adamluzsi/testcase/internal/caller"
	"github.com/adamluzsi/testcase/internal/reflects"
	"github.com/adamluzsi/testcase/pp"
	"path/filepath"
	"reflect"
	"regexp"
	"runtime"
	"strings"
)

type Query struct {
	pkg          string
	receiverType reflect.Type
	fn           struct {
		Name string
		Type reflect.Type
		Func *runtime.Func
	}
}

func (q Query) PackageOf(v any) Query {
	q.pkg = filepath.Base(reflect.TypeOf(v).PkgPath())
	return q
}

func (q Query) Receiver(v any) Query {
	q = q.PackageOf(v)
	q.receiverType = reflects.BaseTypeOf(v)

	return q
}

func (q Query) Package(pkg string) Query {
	q.pkg = pkg
	return q
}

var fnRGX = regexp.MustCompile(`([^.]+)\.\(?([^\)\.])\)?\.([^-]+)-?(?:.*)?$"`)

func (q Query) Function(v any) Query {
	q.fn.Type = reflect.TypeOf(v)
	q.fn.Func = runtime.FuncForPC(reflect.ValueOf(v).Pointer())
	pp.PP(q.fn.Func.Name())
	q.fn.Name = q.fn.Func.Name()

	submatch := fnRGX.FindAllStringSubmatch(q.fn.Name, -1)
	pp.PP(submatch, q.fn.Name)

	return q
}

func (q Query) check() bool {
	return caller.MatchFunc(func(fn caller.Func) bool {
		if !q.isPackage(fn) {
			return false
		}
		if !q.isReceiver(fn) {
			return false
		}
		if !q.isFunction(fn) {
			return false
		}
		return true
	})
}

func (q Query) isPackage(fn caller.Func) bool {
	if q.pkg == "" {
		return true
	}
	if strings.HasSuffix(fn.Package, q.pkg) {
		return true
	}
	return false
}

func (q Query) isReceiver(fn caller.Func) bool {
	if q.receiverType == nil {
		return true
	}
	name := q.receiverType.Name()
	if fn.Receiver == name {
		return true
	}
	if strings.TrimPrefix(fn.Receiver, "*") == name {
		return true
	}
	return false
}

func (q Query) isFunction(fn caller.Func) bool {
	if q.fn.Type == nil {
		return true
	}
	return false
}

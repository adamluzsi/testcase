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

type CallerQuery struct {
	pkg          string
	receiverType reflect.Type
	fn           struct {
		Name string
		Type reflect.Type
		Func *runtime.Func
	}
}

func (q CallerQuery) PackageOf(v any) CallerQuery {
	q.pkg = filepath.Dir(reflect.TypeOf(v).PkgPath())
	return q
}

func (q CallerQuery) Receiver(v any) CallerQuery {
	q.receiverType = reflects.BaseTypeOf(v)
	q.pkg = filepath.Dir(q.receiverType.PkgPath())
	return q
}

func (q CallerQuery) Package(pkg string) CallerQuery {
	q.pkg = pkg
	return q
}

var fnRGX = regexp.MustCompile(`([^.]+)\.\(?([^\)\.])\)?\.([^-]+)-?(?:.*)?$"`)

func (q CallerQuery) Function(v any) CallerQuery {
	q.fn.Type = reflect.TypeOf(v)
	q.fn.Func = runtime.FuncForPC(reflect.ValueOf(v).Pointer())
	pp.PP(q.fn.Func.Name())
	q.fn.Name = q.fn.Func.Name()

	submatch := fnRGX.FindAllStringSubmatch(q.fn.Name, -1)
	pp.PP(submatch, q.fn.Name)

	return q
}

func (q CallerQuery) check() bool {
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

func (q CallerQuery) isPackage(fn caller.Func) bool {
	if q.pkg == "" {
		return true
	}
	if strings.HasSuffix(fn.Package, q.pkg) {
		return true
	}
	return false
}

func (q CallerQuery) isReceiver(fn caller.Func) bool {
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

func (q CallerQuery) isFunction(fn caller.Func) bool {
	if q.fn.Type == nil {
		return true
	}
	return false
}

package faultinject

import (
	"github.com/adamluzsi/testcase/internal/caller"
)

// Tag is a fault Tag which will be matched between the registered tags in Injector and what is injected in the context.Context.
//
// It must be a ~struct type.
type Tag any

// Fault is a special Tag that can inject an Error into fault points (Injector.Check).
type Fault struct {
	Package  string
	Receiver string
	Function string
	Error    error
}

func (ff Fault) check() (error, bool) {
	fn, ok := caller.GetFunc()
	if !ok {
		return nil, false
	}
	if ff.Package != "" && ff.Package != fn.Package {
		return nil, false
	}
	if ff.Receiver != "" && ff.Receiver != fn.Receiver {
		return nil, false
	}
	if ff.Function != "" && ff.Function != fn.Funcion {
		return nil, false
	}
	return ff.Error, true
}

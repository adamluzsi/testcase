package faultinject

import (
	"github.com/adamluzsi/testcase/internal/caller"
)

// CallerFault allows you to inject Fault by Caller stack position.
type CallerFault struct {
	Package  string
	Receiver string
	Function string
}

func (ff CallerFault) check() bool {
	return caller.MatchFunc(func(fn caller.Func) bool {
		if ff.Package != "" && ff.Package != fn.Package {
			return false
		}
		if ff.Receiver != "" && ff.Receiver != fn.Receiver {
			return false
		}
		if ff.Function != "" && ff.Function != fn.Funcion {
			return false
		}
		return true
	})
}

package faultinject

import (
	"strings"

	"go.llib.dev/testcase/internal/caller"
)

// CallerFault allows you to inject Fault by Caller stack position.
type CallerFault struct {
	Package  string
	Receiver string
	Function string
}

func (ff CallerFault) check() bool {
	return caller.MatchFunc(func(fn caller.Func) bool {
		if !ff.isPackage(fn) {
			return false
		}
		if !ff.isReceiver(fn) {
			return false
		}
		if !ff.isFunction(fn) {
			return false
		}
		return true
	})
}

func (ff CallerFault) isPackage(fn caller.Func) bool {
	return ff.Package == "" || ff.Package == fn.Package
}

func (ff CallerFault) isReceiver(fn caller.Func) bool {
	return ff.Receiver == "" ||
		ff.Receiver == fn.Receiver ||
		ff.Receiver == strings.TrimPrefix(ff.Receiver, "*")
}

func (ff CallerFault) isFunction(fn caller.Func) bool {
	return ff.Function == "" || ff.Function == fn.Funcion
}

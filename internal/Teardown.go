package internal

import (
	"fmt"
	"reflect"
	"runtime"
	"sync"
)

type Teardown struct {
	CallerOffset int

	mutex sync.Mutex
	fns   []func()
}

// Defer function defers the execution of a function until the current testCase case returns.
// Deferred functions are guaranteed to run, regardless of panics during the testCase case execution.
// Deferred function calls are pushed onto a testcase runtime stack.
// When an function passed to the Defer function, it will be executed as a deferred call in last-in-first-orderingOutput order.
//
// It is advised to use this inside a testcase.Spec#Let memorization function
// when spec variable defined that has finalizer requirements.
// This allow the specification to ensure the object finalizer requirements to be met,
// without using an testcase.Spec#After where the memorized function would be executed always, regardless of its actual need.
//
// In a practical example, this means that if you have common vars defined with testcase.Spec#Let memorization,
// which needs to be Closed for example, after the testCase case already run.
// Ensuring such objects Close call in an after block would cause an initialization of the memorized object list the time,
// even in tests where this is not needed.
//
// e.g.:
//	- mock initialization with mock controller, where the mock controller #Finish function must be executed after each testCase suite.
//	- sql.DB / sql.Tx
//	- basically anything that has the io.Closer interface
//
func (td *Teardown) Defer(fn interface{}, args ...interface{}) {
	if fn, ok := fn.(func()); ok && len(args) == 0 {
		td.Cleanup(fn)
		return
	}

	rfn := reflect.ValueOf(fn)
	if rfn.Kind() != reflect.Func {
		panic(`T#Defer can only take functions`)
	}
	rfnType := rfn.Type()

	var caller = func() (file string, line int) {
		_, file, line, _ = runtime.Caller(2 + td.CallerOffset)
		return file, line
	}

	if inCount := rfnType.NumIn(); inCount != len(args) {
		file, line := caller()
		const format = "deferred function argument count mismatch: expected %d, but got %d from %s:%d"
		panic(fmt.Sprintf(format, inCount, len(args), file, line))
	}
	var refArgs = make([]reflect.Value, 0, len(args))
	for i, arg := range args {
		value := reflect.ValueOf(arg)
		inType := rfnType.In(i)
		switch expected := inType.Kind(); expected {
		case reflect.Interface:
			if !value.Type().Implements(inType) {
				file, line := caller()
				const format = "deferred function argument[%d] %s doesn't implements %s.%s from %s:%d"
				panic(fmt.Sprintf(format, i, value.Kind(), inType.PkgPath(), inType.Name(), file, line))
			}
		case value.Kind():
			// OK
		default:
			file, line := caller()
			const format = "deferred function argument[%d] type mismatch: expected %s, but got %s from %s:%d"
			panic(fmt.Sprintf(format, i, expected, value.Kind(), file, line))
		}

		refArgs = append(refArgs, value)
	}

	td.Cleanup(func() { rfn.Call(refArgs) })
}

func (td *Teardown) Cleanup(fn func()) {
	td.mutex.Lock()
	defer td.mutex.Unlock()
	td.fns = append(td.fns, func() { InGoroutine(fn) })
}

func (td *Teardown) Finish() {
	for {
		if td.isEmpty() {
			break
		}

		td.run()
	}
}

func (td *Teardown) isEmpty() bool {
	return len(td.fns) == 0
}

func (td *Teardown) run() {
	td.mutex.Lock()
	fns := td.fns
	td.fns = nil
	td.mutex.Unlock()
	for _, cu := range fns {
		defer cu()
	}
}

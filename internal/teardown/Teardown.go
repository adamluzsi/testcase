package teardown

import (
	"fmt"
	"reflect"
	"runtime"
	"sync"

	"go.llib.dev/testcase/sandbox"
)

type Teardown struct {
	CallerOffset int

	mutex sync.Mutex
	fns   []func()
	added chan struct{}
}

// Defer function defers the execution of a function until the current test case returns.
// Deferred functions are guaranteed to run, regardless of panics during the test case execution.
// Deferred function calls are pushed onto a testcase runtime stack.
// When an function passed to the Defer function, it will be executed as a deferred call in last-in-first-orderingOutput order.
//
// It is advised to use this inside a testcase.Spec#Let memorization function
// when spec variable defined that has finalizer requirements.
// This allow the specification to ensure the object finalizer requirements to be met,
// without using an testcase.Spec#After where the memorized function would be executed always, regardless of its actual need.
//
// In a practical example, this means that if you have common vars defined with testcase.Spec#Let memorization,
// which needs to be Closed for example, after the test case already run.
// Ensuring such objects Close call in an after block would cause an initialization of the memorized object list the time,
// even in tests where this is not needed.
//
// e.g.:
//   - mock initialization with mock controller, where the mock controller #Finish function must be executed after each testCase suite.
//   - sql.DB / sql.Tx
//   - basically anything that has the io.Closer interface
//
// https://github.com/golang/go/issues/41891
func (td *Teardown) Defer(fn interface{}, args ...interface{}) {
	if len(args) == 0 {
		switch fn := fn.(type) {
		case func():
			td.add(fn)
			return
		case func() error:
			td.add(func() { _ = fn() })
			return
		}
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

	numInCountMatch := func() bool {
		inCount := rfnType.NumIn()
		if rfnType.IsVariadic() {
			return inCount-1 <= len(args)
		}
		return inCount == len(args)
	}

	getInType := func(index int) reflect.Type {
		if !rfnType.IsVariadic() {
			return rfnType.In(index)
		}
		if index < rfnType.NumIn()-1 {
			return rfnType.In(index)
		}
		return rfnType.In(rfnType.NumIn() - 1).Elem()
	}

	if !numInCountMatch() {
		file, line := caller()
		const format = "deferred function argument count mismatch: expected %d, but got %d from %s:%d"
		panic(fmt.Sprintf(format, rfnType.NumIn(), len(args), file, line))
	}
	var refArgs = make([]reflect.Value, 0, len(args))
	for i, arg := range args {
		value := reflect.ValueOf(arg)
		inType := getInType(i)
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

	td.add(func() { rfn.Call(refArgs) })
}

func (td *Teardown) Finish() {
	for {
		if td.isEmpty() {
			break
		}

		td.run()
	}
}

// FinishUntil keeps the teardown queue drained until the done channel is closed.
//
// Finish returns as soon as the queue is empty, so a function registered after
// that moment is silently dropped. FinishUntil instead stays around and keeps
// draining, which makes it safe for a concurrently running goroutine to register
// a deferred function while it is shutting down. The goroutine's exit may even
// depend on that function running, e.g. when it releases the resource the
// goroutine is blocked on.
//
// FinishUntil must not be called concurrently with Finish or with itself.
func (td *Teardown) FinishUntil(done <-chan struct{}) {
	for {
		td.Finish()

		select {
		case <-done:
			// Nothing can be registered after done is closed by a party that we
			// are waiting for, but something might have been registered while
			// the last drain was still running, so drain one final time.
			td.Finish()
			return
		case <-td.notification():
			// a new function was registered, loop around and run it
		}
	}
}

// notification returns the channel that add signals on.
// A spurious signal is possible, when the registered function was already
// picked up by an in-flight Finish, but a signal is never lost.
func (td *Teardown) notification() <-chan struct{} {
	td.mutex.Lock()
	defer td.mutex.Unlock()
	return td.notify()
}

// notify must be called while td.mutex is held.
func (td *Teardown) notify() chan struct{} {
	if td.added == nil {
		td.added = make(chan struct{}, 1)
	}
	return td.added
}

func (td *Teardown) isEmpty() bool {
	td.mutex.Lock()
	defer td.mutex.Unlock()
	return len(td.fns) == 0
}

func (td *Teardown) add(fn func()) {
	td.mutex.Lock()
	td.fns = append(td.fns, func() { td.recoverGoexit(fn) })
	added := td.notify()
	td.mutex.Unlock()
	select {
	case added <- struct{}{}:
	default: // a pending signal is enough, Finish drains the whole queue
	}
}

func (td *Teardown) recoverGoexit(fn func()) {
	ro := sandbox.Run(fn)
	if ro.OK || ro.Goexit {
		return
	}
	panic(ro.Trace() + "\n")
}

func (td *Teardown) run() {
	td.mutex.Lock()
	fns := td.fns
	td.fns = nil
	td.mutex.Unlock()
	for _, cu := range fns {
		//goland:noinspection GoDeferInLoop
		defer cu()
	}
}

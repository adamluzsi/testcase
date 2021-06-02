package testcase

import (
	"runtime"
	"sync"
)

// Race is a test helper that allows you to create a race situation easily.
// Race will execute each provided anonymous lambda function in a different goroutine,
// and make sure they are scheduled at the same time.
//
// This is useful when you work on a component that requires thread-safety.
// By using the Race helper, you can write an example use of your component,
// and run the testing suite with `go test -race`.
// The race detector then should be able to notice issues with your implementation.
func Race(fn1, fn2 func(), more ...func()) {
	fns := append([]func(){fn1, fn2}, more...)
	var (
		start sync.WaitGroup
		rdy   sync.WaitGroup
		wg    sync.WaitGroup
	)
	start.Add(1) // get ready for the race
	wg.Add(len(fns))
	rdy.Add(len(fns))
	for _, fn := range fns {
		go func(blk func()) {
			defer wg.Done()
			rdy.Done()   // signal that participant is ready
			start.Wait() // line up participants
			blk()
		}(fn)
	}

	runtime.Gosched()
	rdy.Wait()   // wait until everyone lined up
	start.Done() // start the race
	wg.Wait()    // wait members to finish
}

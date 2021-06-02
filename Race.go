package testcase

import (
	"runtime"
	"sync"
)

// Race is a test helper that allows you to create a race situation easily.
//
// This is useful when you work on a component that requires thread-safety.
// By using the Race helper, you can write an example use of your component,
// and run the testing suite with `go test -race`.
// The race detector then should be able to notice issues with your implementation.
func Race(blk func()) int {
	var (
		start sync.WaitGroup
		rdy   sync.WaitGroup
		wg    sync.WaitGroup
		num   = runtime.NumCPU()
	)
	start.Add(1) // get ready for the race

	for i := 0; i < num; i++ {
		wg.Add(1)
		rdy.Add(1)
		go func() {
			defer wg.Done()
			rdy.Done()
			start.Wait() // line up participants
			blk()
		}()
	}

	//rdy.Wait() // wait until everyone lined up
	start.Done() // start the race
	wg.Wait()    // wait members to finish
	return num
}

func Concurrently(fns ...func())  {
	var (
		start sync.WaitGroup
		rdy   sync.WaitGroup
		wg    sync.WaitGroup
		num   = runtime.NumCPU()
	)
	start.Add(1) // get ready for the race

	for i := 0; i < num; i++ {
		wg.Add(1)
		rdy.Add(1)
		go func() {
			defer wg.Done()
			rdy.Done()
			start.Wait() // line up participants
			blk()
		}()
	}

	//rdy.Wait() // wait until everyone lined up
	start.Done() // start the race
	wg.Wait()    // wait members to finish
	return num
}
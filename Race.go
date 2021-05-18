package testcase

import (
	"runtime"
	"sync"
)

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

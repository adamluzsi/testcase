package internal

import "sync"

func InGoroutine(fn func()) {
	var wg sync.WaitGroup
	wg.Add(1)
	var panicValue interface{}
	go func() {
		defer wg.Done()
		defer func() { panicValue = recover() }()
		fn()
	}()
	wg.Wait()

	// This implementation doesn't handle panic(nil).
	// This is intentional because panic(nil) and runtime.Goexit is difficult to differentiate during recovery.
	if panicValue != nil {
		panic(panicValue)
	}
}

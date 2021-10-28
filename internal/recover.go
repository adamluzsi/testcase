package internal

import "sync"

func Recover(fn func()) (panicValue interface{}, ok bool) {
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		defer func() { panicValue = recover() }()
		fn()
		ok = true
	}()
	wg.Wait()
	return
}

func RecoverExceptGoexit(fn func()) {
	panicValue, ok := Recover(fn)
	// This implementation doesn't handle panic(nil).
	// This is intentional because panic(nil) and runtime.Goexit is difficult to differentiate during recovery.
	if !ok && panicValue != nil {
		panic(panicValue)
	}
}

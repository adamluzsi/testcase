package sandbox

import (
	"runtime"
	"sync"

	"github.com/adamluzsi/testcase/internal/caller"
)

func Run(fn func()) (ro RunOutcome) {
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		defer func() { ro.PanicValue = recover() }()
		defer func() { ro.Goexit = stackHasGoexit() }()
		fn()
		ro.OK = true
	}()
	wg.Wait()
	return
}

type RunOutcome struct {
	OK         bool
	PanicValue any
	Goexit     bool
}

func stackHasGoexit() bool {
	const goexitFuncName = "runtime.Goexit"
	return caller.MatchAllFrame(func(frame runtime.Frame) bool {
		return frame.Function == goexitFuncName
	})
}

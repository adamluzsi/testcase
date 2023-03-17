package sandbox

import (
	"bytes"
	"fmt"
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
		defer func() {
			if !ro.OK {
				ro.Frames = getFrames()
			}
		}()
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
	Frames     []runtime.Frame
}

func (ro RunOutcome) Trace() string {
	var buf bytes.Buffer
	switch {
	case ro.Goexit:
		_, _ = buf.Write([]byte("runtime.Goexit"))
	case !ro.OK:
		_, _ = fmt.Fprintf(&buf, "panic: %v", ro.PanicValue)
	}
	_, _ = buf.Write([]byte("\n"))
	for _, frame := range ro.Frames {
		_, _ = fmt.Fprintf(&buf, "%s\n\t%s:%d %#v\n", frame.Function, frame.File, frame.Line, frame.PC)
	}
	return buf.String()
}

func stackHasGoexit() bool {
	const goexitFuncName = "runtime.Goexit"
	return caller.Until(func(frame runtime.Frame) bool {
		return frame.Function == goexitFuncName
	})
}

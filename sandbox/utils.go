package sandbox

import (
	"bytes"
	"fmt"
	"runtime"

	"go.llib.dev/testcase/internal/caller"
)

func (ro O) Trace() string {
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

// OnNotOK will execute the argument block when the OK state is false.
func (ro O) OnNotOK(blk func()) {
	if ro.OK {
		return
	}
	if blk == nil {
		return
	}
	blk()
}

func stackHasGoexit() bool {
	const goexitFuncName = "runtime.Goexit"
	return caller.Until(func(frame runtime.Frame) bool {
		return frame.Function == goexitFuncName
	})
}

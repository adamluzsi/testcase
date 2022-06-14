package caller

import (
	"fmt"
	"path"
	"runtime"
	"strings"
)

var testcasePkgDirPath string

func init() {
	_, specFilePath, _, _ := runtime.Caller(0)                      // this caller
	testcasePkgDirPath = path.Dir(path.Dir(path.Dir(specFilePath))) // ../../../
}

func GetFrame() (_frame runtime.Frame, _ok bool) {
	const maxStackLen = 42
	var pc [maxStackLen]uintptr
	// Skip two extra frames to account for this function
	// and runtime.Callers itself.
	n := runtime.Callers(2, pc[:])
	if n == 0 {
		return runtime.Frame{}, false
	}
	frames := runtime.CallersFrames(pc[:n])
	var firstFrame, frame runtime.Frame
	for more := true; more; {
		frame, more = frames.Next()
		if firstFrame.PC == 0 {
			firstFrame = frame
		}
		if !isValidCallerFile(frame.File) {
			continue
		}
		return frame, true
	}
	// If no "non-helper" frame is found, the first non is frame is returned.
	return firstFrame, true
}

func GetLocation(basename bool) string {
	frame, ok := GetFrame()
	if !ok {
		return ""
	}
	var fname = frame.File
	if basename {
		fname = path.Base(fname)
	}
	return fmt.Sprintf(`%s:%d`, fname, frame.Line)
}

func isValidCallerFile(file string) bool {
	switch {
	// fast path when caller located in a *_test.go file
	case strings.HasSuffix(file, `_test.go`):
		return true
	// skip testcase packages
	case strings.HasPrefix(file, testcasePkgDirPath):
		return false
	// skip stdlib testing
	case strings.Contains(file, `go/src/testing/`):
		return false
	// skip stdlib runtime
	case strings.Contains(file, `go/src/runtime/`):
		return false
	default:
		return true
	}
}

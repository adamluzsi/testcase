package caller

import (
	"fmt"
	"path"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
)

var testcasePkgDirPath string

func init() {
	_, specFilePath, _, _ := runtime.Caller(0)                      // this caller
	testcasePkgDirPath = path.Dir(path.Dir(path.Dir(specFilePath))) // ../../../
}

type Func struct {
	Package  string
	Receiver string
	Funcion  string
}

var rgxGetFuncLambdaSuffix = regexp.MustCompile(`\.func1.*$`)

func GetFunc() (Func, bool) {
	frame, ok := GetFrame()
	if !ok {
		return Func{}, false
	}
	if frame.Function == "" {
		return Func{}, false
	}

	base := filepath.Base(path.Base(frame.Function))
	base = rgxGetFuncLambdaSuffix.ReplaceAllString(base, "")
	fnParts := strings.Split(base, ".")

	var fn Func
	switch len(fnParts) {
	case 3:
		fn.Package = fnParts[0]
		fn.Receiver = strings.Trim(fnParts[1], "()")
		fn.Funcion = fnParts[2]
	case 2:
		fn.Package = fnParts[0]
		fn.Funcion = fnParts[1]
	default:
		return Func{}, false
	}
	return fn, true
}

func GetFrame() (_frame runtime.Frame, _ok bool) {
	return MatchFrame(func(frame runtime.Frame) bool {
		return true
	})
}

func MatchFrame(check func(frame runtime.Frame) bool) (_frame runtime.Frame, _ok bool) {
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
		if !check(frame) {
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

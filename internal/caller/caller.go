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

func convertFrameToFunc(frame runtime.Frame) (Func, bool) {
	if frame.Function == "" {
		return Func{}, false
	}

	base := filepath.Base(path.Base(frame.Function))
	base = rgxGetFuncLambdaSuffix.ReplaceAllString(base, "")
	fnParts := filterGetFuncParts(strings.Split(base, "."))

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

func GetFunc() (Func, bool) {
	frame, ok := GetFrame()
	if !ok {
		return Func{}, false
	}
	return convertFrameToFunc(frame)
}

func MatchFunc(match func(fn Func) bool) bool {
	return MatchFrame(func(frame runtime.Frame) bool {
		fn, ok := convertFrameToFunc(frame)
		if !ok {
			return false
		}
		return match(fn)
	})
}

func filterGetFuncParts(fnParts []string) []string {
	var sliceIndex int
	for i := 0; i < len(fnParts); i++ {
		sliceIndex = i
		if strings.HasPrefix(fnParts[i], "func") {
			break
		}
	}
	return fnParts[:sliceIndex+1]
}

func GetFrame() (frame runtime.Frame, ok bool) {
	MatchFrame(func(frm runtime.Frame) bool {
		frame = frm
		ok = true
		return true
	})
	return
}

func MatchFrame(check func(frame runtime.Frame) bool) bool {
	return MatchAllFrame(isValidCallerFile, check)
}

func MatchAllFrame(checks ...func(frame runtime.Frame) bool) (_ok bool) {
	const maxStackLen = 42
	var pc [maxStackLen]uintptr
	// Skip two extra frames to account for this function
	// and runtime.Callers itself.
	n := runtime.Callers(2, pc[:])
	if n == 0 {
		return false
	}
	frames := runtime.CallersFrames(pc[:n])
	var frame runtime.Frame
seeking:
	for more := true; more; {
		frame, more = frames.Next()
		for _, check := range checks {
			if !check(frame) {
				continue seeking
			}
		}
		return true
	}
	return false
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

func isValidCallerFile(frame runtime.Frame) bool {
	file := frame.File
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

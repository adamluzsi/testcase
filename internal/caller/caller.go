package caller

import (
	"fmt"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
)

var TestcasePkgDirPath string

func init() {
	_, specFilePath, _, _ := runtime.Caller(0)                      // this caller
	TestcasePkgDirPath = path.Dir(path.Dir(path.Dir(specFilePath))) // ../../../
}

type Func struct {
	Package  string
	Receiver string
	Funcion  string
}

func (fn Func) String() string {
	name := fn.Funcion
	if fn.Receiver != "" {
		name = fn.Receiver + "#" + name
	}
	if fn.Package != "" {
		name = fn.Package + "." + name
	}
	return name
}

var rgxGetFuncLambdaSuffix = regexp.MustCompile(`\.func1.*$`)

func FrameToFunc(frame runtime.Frame) (Func, bool) {
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
	return FrameToFunc(frame)
}

func MatchFunc(match func(fn Func) bool) bool {
	return Until(NonTestCaseFrame, func(frame runtime.Frame) bool {
		fn, ok := FrameToFunc(frame)
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
	Until(NonTestCaseFrame, func(frm runtime.Frame) bool {
		frame = frm
		ok = true
		return true
	})
	return
}

func Until(checks ...func(runtime.Frame) bool) (_ok bool) {
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

func GetLocation(asBasename bool) string {
	frame, ok := GetFrame()
	if !ok {
		return ""
	}
	return AsLocation(asBasename, frame.File, frame.Line)
}

func AsLocation(asBasename bool, file string, line int) string {
	if asBasename {
		file = filepath.Base(path.Base(file))
	}
	return fmt.Sprintf(`%s:%d`, file, line)
}

func IsTestFileFrame(frame runtime.Frame) bool {
	return strings.HasSuffix(frame.File, `_test.go`)
}

func IsStdlibFrame(frame runtime.Frame) bool {
	if root, ok := os.LookupEnv("GOROOT"); ok && filepath.IsAbs(root) {
		return strings.Contains(frame.File, root)
	}
	return false
}

func SkipFrame(n int) func(frame runtime.Frame) bool {
	return func(frame runtime.Frame) bool {
		if n <= 0 {
			return true
		}
		n--
		return false
	}
}

func NonTestCaseFrame(frame runtime.Frame) bool {
	file := frame.File
	switch {
	// fast path when caller located in a *_test.go file
	case IsTestFileFrame(frame):
		return true
	// skip testcase/internal packages
	case strings.HasPrefix(file, filepath.Join(TestcasePkgDirPath, "internal")):
		return false
	// skip top level testcase package
	case filepath.Dir(file) == TestcasePkgDirPath:
		return false
	// skip stdlib
	case IsStdlibFrame(frame):
		return false
	default:
		return true
	}
}

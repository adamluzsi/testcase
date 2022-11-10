package sandbox

import (
	"path"
	"runtime"
	"strings"

	"github.com/adamluzsi/testcase/internal/caller"
)

func getFrames() (frames []runtime.Frame) {
	caller.Until(caller.NonTestCaseFrame, isNotSandboxPkg, func(frame runtime.Frame) bool {
		frames = append(frames, frame)
		return false
	})
	return
}

var pkgDir string

func init() {
	_, filePath, _, _ := runtime.Caller(0) // this caller
	pkgDir = path.Dir(filePath)
}

func isNotSandboxPkg(frame runtime.Frame) bool {
	switch {
	case caller.IsTestFileFrame(frame):
		return true
	case strings.HasPrefix(frame.File, pkgDir):
		return false
	default:
		return true
	}
}

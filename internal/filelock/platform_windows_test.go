//go:build windows

package filelock_test

import (
	"go.llib.dev/testcase"
	"go.llib.dev/testcase/internal/filelock"
)

// Windows has no file permissions or groups to describe beyond the shared specs.

func specTryLockOnPlatform(*testcase.Spec, testcase.Var[string], func(*testcase.T) (*filelock.Lock, bool, error)) {
}

func specMkdirGroupOnPlatform(*testcase.Spec, testcase.Var[string], testcase.Var[string], func(*testcase.T) error) {
}

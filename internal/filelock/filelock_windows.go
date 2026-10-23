//go:build windows

package filelock

import (
	"io/fs"
	"os"
	"syscall"
	"unsafe"
)

const noFollow = 0

func createLockFile(path string) error {
	file, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE|os.O_EXCL, lockFilePerm)
	if err != nil {
		return err
	}
	return file.Close()
}

// shareWithGroup has nothing to do, as Windows doesn't have POSIX groups.
func shareWithGroup(string, fs.FileInfo, fs.FileInfo) error { return nil }

var (
	modkernel32      = syscall.NewLazyDLL("kernel32.dll")
	procLockFileEx   = modkernel32.NewProc("LockFileEx")
	procUnlockFileEx = modkernel32.NewProc("UnlockFileEx")
)

const (
	lockfileFailImmediately = 0x00000001
	lockfileExclusiveLock   = 0x00000002

	errorLockViolation syscall.Errno = 33
)

// lockedRange is the byte range that the lock covers.
// Windows doesn't let others read a locked range, so the range starts far beyond the holder data,
// which keeps the holder data readable for ReadHolder.
// Windows allows locking a range beyond the end of the file.
func lockedRange() syscall.Overlapped {
	return syscall.Overlapped{OffsetHigh: 1 << 30}
}

func tryLockFile(file *os.File) (bool, error) {
	ol := lockedRange()
	r1, _, err := procLockFileEx.Call(file.Fd(),
		uintptr(lockfileExclusiveLock|lockfileFailImmediately), 0,
		1, 0, uintptr(unsafe.Pointer(&ol)))
	if r1 != 0 {
		return true, nil
	}
	if err == errorLockViolation {
		return false, nil
	}
	return false, &os.PathError{Op: "LockFileEx", Path: file.Name(), Err: err}
}

// unlockFile is best effort, as closing the file releases the lock anyway.
func unlockFile(file *os.File) {
	ol := lockedRange()
	_, _, _ = procUnlockFileEx.Call(file.Fd(), 0,
		1, 0, uintptr(unsafe.Pointer(&ol)))
}

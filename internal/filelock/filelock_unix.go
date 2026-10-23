//go:build darwin || dragonfly || freebsd || linux || netbsd || openbsd

package filelock

import (
	"os"
	"path/filepath"
	"syscall"
)

const noFollow = syscall.O_NOFOLLOW

// createLockFile creates the lock file with its final permissions in a single step.
// Creating the file directly at the path would let others see it
// with the permissions that the umask left, before a chmod could fix them.
// So the file is prepared under a temporary name, and then linked to the path,
// which fails when the path already exists.
func createLockFile(path string) error {
	tmp, err := os.CreateTemp(filepath.Dir(path), "."+filepath.Base(path)+".*")
	if err != nil {
		return err
	}
	defer os.Remove(tmp.Name())
	defer tmp.Close()
	// The chmod only grants access to the group, so if it fails,
	// the group members get a permission error, rather than more access than intended.
	_ = tmp.Chmod(lockFilePerm)
	return os.Link(tmp.Name(), path)
}

// tryLockFile doesn't have to retry on EINTR, as a non-blocking flock doesn't wait for anything to be interrupted.
func tryLockFile(file *os.File) (bool, error) {
	err := syscall.Flock(int(file.Fd()), syscall.LOCK_EX|syscall.LOCK_NB)
	if err == syscall.EWOULDBLOCK {
		return false, nil
	}
	return err == nil, os.NewSyscallError("flock", err)
}

func unlockFile(file *os.File) {
	_ = syscall.Flock(int(file.Fd()), syscall.LOCK_UN)
}

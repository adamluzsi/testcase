//go:build darwin || dragonfly || freebsd || linux || netbsd || openbsd || windows

// Package filelock implements an exclusive, host-wide lock that is backed by a lock file.
//
// The lock relies on the operating system's advisory file locking,
// which the operating system releases when the holder process exits,
// so a crashed or killed process can't leave a stale lock behind.
//
// On platforms without file locking support, the lock is always available,
// and the package doesn't touch the file system.
package filelock

import (
	"errors"
	"io/fs"
	"os"
)

// Lock is an acquired lock file.
type Lock struct {
	path string
	file *os.File
}

// TryLock attempts to acquire the lock file at the given path without blocking.
// It reports false when someone else holds the lock, or has just released it.
// The directory of the path has to exist already, see MkdirGroup.
//
// The data is written into the lock file to describe the current holder,
// so others can learn about it with ReadHolder.
//
// The lock file is readable and writable by the group of its directory,
// so the permissions of the directory decide who can use the lock.
// TryLock never acquires the lock through a symbolic link at the path, nor does it write the link's target.
func TryLock(path string, data []byte) (*Lock, bool, error) {
	file, err := openLockFile(path)
	if err != nil {
		return nil, false, err
	}
	ok, err := tryLockFile(file)
	// The previous holder removes the lock file before it releases the lock,
	// so the file might no longer be the lock file at the path by the time we lock it.
	// That is reported like a lock that someone else holds, and the next attempt opens the current lock file.
	if err != nil || !ok || !isCurrent(file, path) {
		_ = file.Close()
		return nil, false, err
	}
	// The data only describes the holder for the others, so failing to write it doesn't fail the lock.
	_ = file.Truncate(0)
	_, _ = file.WriteAt(data, 0)
	return &Lock{path: path, file: file}, true, nil
}

// lockFilePerm lets the members of the directory's group use the lock file too.
const lockFilePerm = 0o660

// openLockFile opens the lock file at the path, and creates it first when it doesn't exist yet.
func openLockFile(path string) (*os.File, error) {
	for {
		file, err := os.OpenFile(path, os.O_RDWR|noFollow, 0)
		if !errors.Is(err, fs.ErrNotExist) {
			return file, err
		}
		// someone else might create the lock file in the meantime, which is just as good
		if err := createLockFile(path); err != nil && !errors.Is(err, fs.ErrExist) {
			return nil, err
		}
	}
}

// isCurrent reports whether the file is still the lock file at the path.
func isCurrent(file *os.File, path string) bool {
	locked, lockedErr := file.Stat()
	current, currentErr := os.Lstat(path)
	return lockedErr == nil && currentErr == nil && os.SameFile(locked, current)
}

// Unlock releases the lock and removes the lock file.
//
// Unlock can't fail: closing the lock file releases the lock even when Close reports an error.
func (l *Lock) Unlock() {
	if l.file == nil {
		return
	}
	// Removal happens while the lock is still held, so anyone who
	// locks the old file afterwards notices that it is no longer current.
	removeErr := os.Remove(l.path)
	unlockFile(l.file)
	_ = l.file.Close()
	l.file = nil
	if removeErr != nil && !errors.Is(removeErr, fs.ErrNotExist) {
		// Windows doesn't allow removing an open file, so removal is attempted again once the file is closed.
		// It is still safe, as Windows refuses the removal while someone else has the file open too,
		// so it can only remove a lock file that nobody uses.
		// Elsewhere, the removal fails for the same reason again, and the lock file just remains for the next holder.
		_ = os.Remove(l.path)
	}
}

// ReadHolder returns the data that the current holder of the lock file wrote into it.
// It returns false if the holder data is not available.
func ReadHolder(path string) (string, bool) {
	data, err := os.ReadFile(path)
	return string(data), err == nil && 0 < len(data)
}

//go:build darwin || dragonfly || freebsd || linux || netbsd || openbsd || windows

package filelock

import (
	"errors"
	"io/fs"
	"os"
	"syscall"
)

// MkdirGroup makes sure that path is a directory that only its owner and the group of the groupOf path can access.
// The entries that the group members create in the directory belong to the group too,
// so the lock files of the directory are shared among the group members.
// The parent directory of the path has to exist already.
//
// When the current user can't give the directory to the group, e.g. because it isn't a member of the group,
// the directory stays private to the current user.
// When someone else owns the directory, MkdirGroup only accepts it if it belongs to the group,
// and others can't write it. MkdirGroup refuses a symbolic link at the path.
//
// Windows doesn't have POSIX groups, so there the directory keeps the permissions that it inherits from its parent.
func MkdirGroup(path, groupOf string) error {
	group, err := os.Stat(groupOf)
	if err != nil {
		return err
	}
	dir, err := mkdir(path)
	if err != nil {
		return err
	}
	return shareWithGroup(path, dir, group)
}

const privateDirMode = 0o700

// mkdir makes sure that path is a directory.
// It creates the directory as private, as its permissions are only set afterwards.
func mkdir(path string) (fs.FileInfo, error) {
	if err := os.Mkdir(path, privateDirMode); err != nil && !errors.Is(err, fs.ErrExist) {
		return nil, err
	}
	info, err := os.Lstat(path)
	// On Windows, a symbolic link to a directory reports both ModeSymlink and ModeDir,
	// so only the exact type tells a real directory apart.
	if err == nil && info.Mode().Type() != fs.ModeDir {
		err = &os.PathError{Op: "mkdir", Path: path, Err: syscall.ENOTDIR}
	}
	return info, err
}

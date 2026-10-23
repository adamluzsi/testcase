//go:build darwin || dragonfly || freebsd || linux || netbsd || openbsd

package filelock

import (
	"fmt"
	"io/fs"
	"os"
	"syscall"
)

// The setgid bit makes the new entries of the directory inherit the directory's group.
const groupDirMode = fs.ModeSetgid | 0o770

// shareWithGroup gives the directory to the group, and lets only the owner and the group access it.
func shareWithGroup(path string, dir, group fs.FileInfo) error {
	var (
		gid    = gidOf(group)
		dirGID = gidOf(dir)
	)
	if !isOwner(dir) {
		switch {
		case dirGID != gid:
			return fmt.Errorf("%s is owned by another user, and it belongs to group %d instead of %d", path, dirGID, gid)
		case dir.Mode()&0o002 != 0:
			return fmt.Errorf("%s is owned by another user, and everyone can write it", path)
		}
		return nil
	}
	mode := groupDirMode
	if err := os.Lchown(path, -1, gid); err != nil {
		// The directory can't be given to the group, e.g. because the current user isn't a member of the group,
		// so it stays private to the current user.
		mode = privateDirMode
	}
	// Changing the group might clear the setgid bit, so the permissions are set afterwards.
	return os.Chmod(path, mode)
}

func gidOf(info fs.FileInfo) int { return int(info.Sys().(*syscall.Stat_t).Gid) }

func isOwner(info fs.FileInfo) bool {
	return int(info.Sys().(*syscall.Stat_t).Uid) == os.Geteuid()
}

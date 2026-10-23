//go:build darwin || dragonfly || freebsd || linux || netbsd || openbsd

package filelock_test

import (
	"io/fs"
	"os"
	"syscall"

	"go.llib.dev/testcase"
	"go.llib.dev/testcase/assert"
)

// specMkdirGroupOnPlatform describes the behaviour of MkdirGroup that relies on Unix file permissions and groups.
func specMkdirGroupOnPlatform(s *testcase.Spec, path, groupOf testcase.Var[string], act func(*testcase.T) error) {
	s.Then("the directory is only accessible to its owner and the group", func(t *testcase.T) {
		assert.NoError(t, act(t))

		assert.Equal(t, fileMode(t, path.Get(t)), fs.ModeDir|fs.ModeSetgid|0o770)
	})

	s.Then("the directory belongs to the group of the groupOf path", func(t *testcase.T) {
		assert.NoError(t, act(t))

		assert.Equal(t, fileGID(t, path.Get(t)), fileGID(t, groupOf.Get(t)))
	})

	s.When("the umask strips every permission of the group and others", func(s *testcase.Spec) {
		s.Before(func(t *testcase.T) { restrictUmask(t) })

		s.Then("the directory is still accessible to the group", func(t *testcase.T) {
			assert.NoError(t, act(t))

			assert.Equal(t, fileMode(t, path.Get(t)), fs.ModeDir|fs.ModeSetgid|0o770)
		})
	})

	s.When("a private directory already exists at the path", func(s *testcase.Spec) {
		s.Before(func(t *testcase.T) {
			assert.Must(t).NoError(os.Mkdir(path.Get(t), 0o700))
		})

		s.Then("it makes the directory accessible to the group", func(t *testcase.T) {
			assert.NoError(t, act(t))

			assert.Equal(t, fileMode(t, path.Get(t)), fs.ModeDir|fs.ModeSetgid|0o770)
		})
	})

	s.When("the groupOf path belongs to another group of the current user", func(s *testcase.Spec) {
		s.Before(func(t *testcase.T) {
			gid, ok := otherGroupOf(t, groupOf.Get(t))
			if !ok {
				t.Skip("the current user is not a member of any other group")
			}
			assert.Must(t).NoError(os.Lchown(groupOf.Get(t), -1, gid))
		})

		s.Then("the directory belongs to that group too", func(t *testcase.T) {
			assert.NoError(t, act(t))

			assert.Equal(t, fileGID(t, path.Get(t)), fileGID(t, groupOf.Get(t)))
			assert.Equal(t, fileMode(t, path.Get(t)), fs.ModeDir|fs.ModeSetgid|0o770)
		})
	})

	s.When("the current user is not a member of the groupOf path's group", func(s *testcase.Spec) {
		groupOf.Let(s, func(t *testcase.T) string {
			if os.Geteuid() == 0 {
				t.Skip("the super user can give a directory to any group")
			}
			for _, p := range []string{"/", "/etc", "/usr", "/usr/bin", "/var", "/sbin"} {
				if !isGroupMember(t, fileGID(t, p)) {
					return p
				}
			}
			t.Skip("no path was found with a group that the current user is not a member of")
			return ""
		})

		s.Then("the directory stays private to the current user", func(t *testcase.T) {
			assert.NoError(t, act(t))

			assert.Equal(t, fileMode(t, path.Get(t)), fs.ModeDir|0o700)
		})
	})

	s.When("another user owns the directory", func(s *testcase.Spec) {
		path.Let(s, func(t *testcase.T) string {
			return systemDir(t, func(mode fs.FileMode) bool { return mode&0o002 == 0 }, "/usr", "/bin", "/sbin")
		})
		groupOf.Let(s, func(t *testcase.T) string {
			return path.Get(t)
		})

		s.Then("it accepts the directory without changing it", func(t *testcase.T) {
			before := fileMode(t, path.Get(t))

			assert.NoError(t, act(t))
			assert.Equal(t, fileMode(t, path.Get(t)), before)
		})

		s.And("the directory doesn't belong to the group of the groupOf path", func(s *testcase.Spec) {
			groupOf.Let(s, func(t *testcase.T) string {
				dir := t.TempDir()
				if fileGID(t, dir) == fileGID(t, path.Get(t)) {
					t.Skip("the new directories of the current user belong to the group of the system directory")
				}
				return dir
			})

			s.Then("it refuses the directory", func(t *testcase.T) {
				assert.Error(t, act(t))
			})
		})

		s.And("everyone can write the directory", func(s *testcase.Spec) {
			path.Let(s, func(t *testcase.T) string {
				return systemDir(t, func(mode fs.FileMode) bool { return mode&0o002 != 0 }, "/var/tmp", "/tmp")
			})

			s.Then("it refuses the directory", func(t *testcase.T) {
				assert.Error(t, act(t))
			})
		})
	})
}

// systemDir returns the first of the paths that is a directory of another user, and whose permissions match.
// A test can't make a directory that another user owns, so the system's directories stand in for one.
// MkdirGroup doesn't change a directory of another user, and without super user privileges, it couldn't either.
func systemDir(t *testcase.T, match func(fs.FileMode) bool, paths ...string) string {
	if os.Geteuid() == 0 {
		t.Skip("the super user owns the system directories, and could change them")
	}
	for _, p := range paths {
		info, err := os.Lstat(p)
		if err == nil && info.IsDir() && int(info.Sys().(*syscall.Stat_t).Uid) != os.Geteuid() && match(info.Mode()) {
			return p
		}
	}
	t.Skip("none of the system directories are suitable:", paths)
	return ""
}

func fileGID(t *testcase.T, path string) int {
	info, err := os.Lstat(path)
	assert.Must(t).NoError(err)
	return int(info.Sys().(*syscall.Stat_t).Gid)
}

func isGroupMember(t *testcase.T, gid int) bool {
	if gid == os.Getegid() {
		return true
	}
	groups, err := os.Getgroups()
	assert.Must(t).NoError(err)
	for _, g := range groups {
		if g == gid {
			return true
		}
	}
	return false
}

// otherGroupOf finds a group of the current user, other than the group of the path.
func otherGroupOf(t *testcase.T, path string) (int, bool) {
	groups, err := os.Getgroups()
	assert.Must(t).NoError(err)
	for _, gid := range append(groups, os.Getegid()) {
		if gid != fileGID(t, path) {
			return gid, true
		}
	}
	return 0, false
}

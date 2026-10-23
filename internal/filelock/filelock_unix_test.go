//go:build darwin || dragonfly || freebsd || linux || netbsd || openbsd

package filelock_test

import (
	"io/fs"
	"os"
	"path/filepath"
	"syscall"

	"go.llib.dev/testcase"
	"go.llib.dev/testcase/assert"
	"go.llib.dev/testcase/internal/filelock"
)

// specTryLockOnPlatform describes the behaviour of TryLock that relies on Unix file permissions.
func specTryLockOnPlatform(s *testcase.Spec, path testcase.Var[string], act func(*testcase.T) (*filelock.Lock, bool, error)) {
	s.Then("the lock file is readable and writable by the group of its directory", func(t *testcase.T) {
		_, _, err := act(t)
		assert.NoError(t, err)

		assert.Equal(t, fileMode(t, path.Get(t)), fs.FileMode(0o660))
	})

	s.When("the umask strips every permission of the group and others", func(s *testcase.Spec) {
		s.Before(func(t *testcase.T) { restrictUmask(t) })

		s.Then("the lock file is still readable and writable by the group of its directory", func(t *testcase.T) {
			_, _, err := act(t)
			assert.NoError(t, err)

			assert.Equal(t, fileMode(t, path.Get(t)), fs.FileMode(0o660))
		})
	})

	s.When("the lock file can't be removed", func(s *testcase.Spec) {
		s.Before(func(t *testcase.T) {
			dir := filepath.Dir(path.Get(t))
			assert.Must(t).NoError(os.WriteFile(path.Get(t), nil, 0o600))
			assert.Must(t).NoError(os.Chmod(dir, 0o500))
			t.Cleanup(func() { _ = os.Chmod(dir, 0o700) })
		})

		s.Then("unlocking still releases the lock", func(t *testcase.T) {
			lock, ok, err := act(t)
			assert.Must(t).NoError(err)
			assert.Must(t).True(ok)

			lock.Unlock()
			next, ok, err := filelock.TryLock(path.Get(t), nil)
			assert.NoError(t, err)
			assert.True(t, ok)
			next.Unlock()
		})
	})
}

// restrictUmask makes the umask strip every permission of the group and others,
// so the tests show that the permissions are set regardless of the umask.
func restrictUmask(t *testcase.T) {
	old := syscall.Umask(0o077)
	t.Cleanup(func() { syscall.Umask(old) })
}

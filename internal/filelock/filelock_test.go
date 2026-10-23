//go:build darwin || dragonfly || freebsd || linux || netbsd || openbsd || windows

package filelock_test

import (
	"io/fs"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"go.llib.dev/testcase"
	"go.llib.dev/testcase/assert"
	"go.llib.dev/testcase/internal/filelock"
	"go.llib.dev/testcase/let"
)

func TestTryLock(t *testing.T) {
	s := testcase.NewSpec(t)

	var (
		path = let.Var(s, func(t *testcase.T) string {
			return filepath.Join(t.TempDir(), "test.lock")
		})
		data = let.Var(s, func(t *testcase.T) []byte {
			return []byte(t.Random.String())
		})
	)
	act := func(t *testcase.T) (*filelock.Lock, bool, error) {
		lock, ok, err := filelock.TryLock(path.Get(t), data.Get(t))
		if ok {
			t.Cleanup(lock.Unlock)
		}
		return lock, ok, err
	}

	s.Then("it acquires the lock", func(t *testcase.T) {
		_, ok, err := act(t)
		assert.NoError(t, err)

		assert.True(t, ok)
	})

	s.Then("the holder's data is available to others", func(t *testcase.T) {
		_, _, err := act(t)
		assert.NoError(t, err)

		got, ok := filelock.ReadHolder(path.Get(t))
		assert.True(t, ok)
		assert.Equal(t, got, string(data.Get(t)))
	})

	s.Then("unlocking removes the lock file", func(t *testcase.T) {
		lock, _, err := act(t)
		assert.NoError(t, err)

		lock.Unlock()
		_, err = os.Lstat(path.Get(t))
		assert.ErrorIs(t, err, fs.ErrNotExist)
	})

	s.When("someone else holds the lock", func(s *testcase.Spec) {
		other := let.Var(s, func(t *testcase.T) *filelock.Lock {
			lock, ok, err := filelock.TryLock(path.Get(t), []byte("other"))
			assert.Must(t).NoError(err)
			assert.Must(t).True(ok)
			t.Cleanup(lock.Unlock)
			return lock
		}).EagerLoading(s)

		s.Then("it reports that the lock is taken", func(t *testcase.T) {
			_, ok, err := act(t)
			assert.NoError(t, err)

			assert.False(t, ok)
		})

		s.Then("it acquires the lock once the other holder releases it", func(t *testcase.T) {
			other.Get(t).Unlock()

			_, ok, err := act(t)
			assert.NoError(t, err)
			assert.True(t, ok)
		})
	})

	s.When("a symbolic link is at the path", func(s *testcase.Spec) {
		target := let.Var(s, func(t *testcase.T) string {
			p := filepath.Join(t.TempDir(), "target")
			assert.Must(t).NoError(os.WriteFile(p, []byte("content"), 0o600))
			return p
		})
		s.Before(func(t *testcase.T) {
			symlink(t, target.Get(t), path.Get(t))
		})

		s.Then("it doesn't acquire the lock through the link", func(t *testcase.T) {
			_, ok, _ := act(t)

			assert.False(t, ok)
		})

		s.Then("it leaves the link's target untouched", func(t *testcase.T) {
			_, _, _ = act(t)

			got, err := os.ReadFile(target.Get(t))
			assert.Must(t).NoError(err)
			assert.Equal(t, string(got), "content")
		})
	})

	s.When("the directory of the path doesn't exist", func(s *testcase.Spec) {
		path.Let(s, func(t *testcase.T) string {
			return filepath.Join(t.TempDir(), "missing", "test.lock")
		})

		s.Then("it fails", func(t *testcase.T) {
			_, _, err := act(t)

			assert.ErrorIs(t, err, fs.ErrNotExist)
		})
	})

	specTryLockOnPlatform(s, path, act)
}

func TestReadHolder(t *testing.T) {
	s := testcase.NewSpec(t)

	var (
		path = let.Var(s, func(t *testcase.T) string {
			return filepath.Join(t.TempDir(), "test.lock")
		})
		data = let.Var(s, func(t *testcase.T) []byte {
			return []byte(t.Random.String())
		})
		holder = let.Var(s, func(t *testcase.T) *filelock.Lock {
			lock, ok, err := filelock.TryLock(path.Get(t), data.Get(t))
			assert.Must(t).NoError(err)
			assert.Must(t).True(ok)
			t.Cleanup(lock.Unlock)
			return lock
		}).EagerLoading(s)
	)
	act := func(t *testcase.T) (string, bool) {
		return filelock.ReadHolder(path.Get(t))
	}

	s.Then("it returns the data that the holder wrote", func(t *testcase.T) {
		got, ok := act(t)

		assert.True(t, ok)
		assert.Equal(t, got, string(data.Get(t)))
	})

	s.When("the holder wrote no data", func(s *testcase.Spec) {
		data.Let(s, func(t *testcase.T) []byte { return nil })

		s.Then("it reports that the holder data is not available", func(t *testcase.T) {
			_, ok := act(t)

			assert.False(t, ok)
		})
	})

	s.When("the holder released the lock", func(s *testcase.Spec) {
		s.Before(func(t *testcase.T) {
			holder.Get(t).Unlock()
		})

		s.Then("it reports that the holder data is not available", func(t *testcase.T) {
			_, ok := act(t)

			assert.False(t, ok)
		})
	})
}

func fileMode(t *testcase.T, path string) fs.FileMode {
	info, err := os.Lstat(path)
	assert.Must(t).NoError(err)
	return info.Mode()
}

// symlink creates a symbolic link at the link path.
// Windows only lets privileged users, or users in developer mode, create symbolic links,
// so there the test is skipped when the link can't be created.
func symlink(t *testcase.T, target, link string) {
	err := os.Symlink(target, link)
	if err != nil && runtime.GOOS == "windows" {
		t.Skip("unable to create a symbolic link:", err)
	}
	assert.Must(t).NoError(err)
}

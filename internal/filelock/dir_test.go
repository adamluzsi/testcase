//go:build darwin || dragonfly || freebsd || linux || netbsd || openbsd || windows

package filelock_test

import (
	"io/fs"
	"os"
	"path/filepath"
	"testing"

	"go.llib.dev/testcase"
	"go.llib.dev/testcase/assert"
	"go.llib.dev/testcase/internal/filelock"
	"go.llib.dev/testcase/let"
)

func TestMkdirGroup(t *testing.T) {
	s := testcase.NewSpec(t)

	var (
		groupOf = let.Var(s, func(t *testcase.T) string {
			return t.TempDir()
		})
		path = let.Var(s, func(t *testcase.T) string {
			return filepath.Join(t.TempDir(), "group")
		})
	)
	act := func(t *testcase.T) error {
		return filelock.MkdirGroup(path.Get(t), groupOf.Get(t))
	}

	s.Then("it creates the directory", func(t *testcase.T) {
		assert.NoError(t, act(t))

		assert.True(t, fileMode(t, path.Get(t)).IsDir())
	})

	s.When("the directory already exists", func(s *testcase.Spec) {
		s.Before(func(t *testcase.T) {
			assert.Must(t).NoError(os.Mkdir(path.Get(t), 0o700))
		})

		s.Then("it accepts the directory", func(t *testcase.T) {
			assert.NoError(t, act(t))
		})
	})

	s.When("a file is at the path", func(s *testcase.Spec) {
		s.Before(func(t *testcase.T) {
			assert.Must(t).NoError(os.WriteFile(path.Get(t), []byte("content"), 0o600))
		})

		s.Then("it fails", func(t *testcase.T) {
			assert.Error(t, act(t))
		})
	})

	s.When("a symbolic link is at the path", func(s *testcase.Spec) {
		target := let.Var(s, func(t *testcase.T) string {
			return t.TempDir()
		})
		s.Before(func(t *testcase.T) {
			symlink(t, target.Get(t), path.Get(t))
		})

		s.Then("it fails without touching the link's target", func(t *testcase.T) {
			before := fileMode(t, target.Get(t))

			assert.Error(t, act(t))
			assert.Equal(t, fileMode(t, target.Get(t)), before)
		})
	})

	s.When("the parent directory doesn't exist", func(s *testcase.Spec) {
		path.Let(s, func(t *testcase.T) string {
			return filepath.Join(t.TempDir(), "missing", "group")
		})

		s.Then("it fails", func(t *testcase.T) {
			assert.ErrorIs(t, act(t), fs.ErrNotExist)
		})
	})

	s.When("the groupOf path doesn't exist", func(s *testcase.Spec) {
		groupOf.Let(s, func(t *testcase.T) string {
			return filepath.Join(t.TempDir(), "missing")
		})

		s.Then("it fails", func(t *testcase.T) {
			assert.ErrorIs(t, act(t), fs.ErrNotExist)
		})
	})

	specMkdirGroupOnPlatform(s, path, groupOf, act)
}

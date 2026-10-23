package testcase_test

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"testing"
	"time"

	"go.llib.dev/testcase"
	"go.llib.dev/testcase/assert"
	"go.llib.dev/testcase/let"
	"go.llib.dev/testcase/sandbox"
)

func ExampleSync() {
	var t *testing.T

	t.Run("foo", func(t *testing.T) {
		t.Parallel()
		// Only a single test can hold the "database" key at a time,
		// including the tests of other packages, which run as separate test processes.
		testcase.Sync(t, "database")
		// From here, the test can use the database exclusively.
	})

	t.Run("bar", func(t *testing.T) {
		t.Parallel()
		testcase.Sync(t, "database") // waits until "foo" finishes
	})
}

func ExampleSync_spec() {
	var tb testing.TB
	s := testcase.NewSpec(tb)

	s.Before(func(t *testcase.T) {
		testcase.Sync(t, "database")
	})

	s.Test("uses the database exclusively", func(t *testcase.T) {})
}

const (
	// syncBlockingTimeout is how long a Sync call has to wait to consider it blocked.
	syncBlockingTimeout = 100 * time.Millisecond
	syncTimeout         = 10 * time.Second
)

func TestSync(t *testing.T) {
	s := testcase.NewSpec(t)

	var (
		// a unique key ensures that the concurrently running test processes don't affect the test
		key = let.Var(s, func(t *testcase.T) string {
			return "TestSync-" + t.Random.UUID()
		})
		subject = let.Var(s, func(t *testcase.T) *testcase.FakeTB {
			return newSyncTB(t, "TestSubject")
		})
	)
	act := func(t *testcase.T) {
		testcase.Sync(subject.Get(t), key.Get(t))
	}

	s.Then("the key is acquired without waiting", func(t *testcase.T) {
		assert.Within(t, syncTimeout, func(context.Context) { act(t) })
	})

	s.Then("other tests wait for the key until the test finishes", func(t *testcase.T) {
		act(t)
		other := newSyncTB(t, "TestOther")

		w := assert.NotWithin(t, syncBlockingTimeout, func(context.Context) {
			testcase.Sync(other, key.Get(t))
		})

		subject.Get(t).Finish()
		assert.Within(t, syncTimeout, func(context.Context) { w.Wait() })
	})

	s.Then("cleanups registered after Sync run while the test still holds the key", func(t *testcase.T) {
		act(t)
		other := newSyncTB(t, "TestOther")

		var w *assert.Async
		subject.Get(t).Cleanup(func() {
			w = assert.NotWithin(t, syncBlockingTimeout, func(context.Context) {
				testcase.Sync(other, key.Get(t))
			})
		})

		subject.Get(t).Finish()
		assert.Within(t, syncTimeout, func(context.Context) { w.Wait() })
	})

	s.Then("other keys can be acquired while the test holds the key", func(t *testcase.T) {
		act(t)
		var keys []string
		for i := 0; i < 3; i++ {
			keys = append(keys, key.Get(t)+"-"+t.Random.UUID())
		}

		assert.Within(t, syncTimeout, func(context.Context) {
			for _, k := range keys {
				testcase.Sync(subject.Get(t), k)
			}
		})
	})

	s.Then("the lock file of the key is shared with the users of the project, through the project's group", func(t *testcase.T) {
		if runtime.GOOS == "windows" {
			t.Skip("Windows doesn't have Unix file permissions")
		}
		act(t)

		var (
			lockFile   = findSyncLockFile(t, key.Get(t))
			projectDir = filepath.Dir(lockFile)
		)
		assert.Equal(t, fileMode(t, projectDir), fs.ModeDir|fs.ModeSetgid|0o770)
		assert.Equal(t, fileMode(t, lockFile), fs.FileMode(0o660))
	})

	s.When("the key is empty", func(s *testcase.Spec) {
		key.LetValue(s, "")

		s.Then("the test fails", func(t *testcase.T) {
			out := sandbox.Run(func() { act(t) })

			assert.False(t, out.OK)
			assert.True(t, subject.Get(t).IsFailed)
		})
	})

	s.When("the test already holds the key", func(s *testcase.Spec) {
		s.Before(func(t *testcase.T) {
			testcase.Sync(subject.Get(t), key.Get(t))
		})

		s.Then("it doesn't wait", func(t *testcase.T) {
			assert.Within(t, syncTimeout, func(context.Context) { act(t) })
		})

		s.Then("the key is released when the test finishes", func(t *testcase.T) {
			act(t)
			other := newSyncTB(t, "TestOther")

			subject.Get(t).Finish()
			assert.Within(t, syncTimeout, func(context.Context) {
				testcase.Sync(other, key.Get(t))
			})
		})
	})

	s.When("another test holds the key", func(s *testcase.Spec) {
		var (
			holder = let.Var(s, func(t *testcase.T) *testcase.FakeTB {
				return newSyncTB(t, "TestHolder")
			})
			holderKey = let.Var(s, func(t *testcase.T) string {
				return key.Get(t)
			})
		)
		s.Before(func(t *testcase.T) {
			testcase.Sync(holder.Get(t), holderKey.Get(t))
		})

		s.Then("it waits until the other test finishes", func(t *testcase.T) {
			w := assert.NotWithin(t, syncBlockingTimeout, func(context.Context) { act(t) })

			holder.Get(t).Finish()
			assert.Within(t, syncTimeout, func(context.Context) { w.Wait() })
		})

		s.Then("the waiting is logged along with the name of the holder test", func(t *testcase.T) {
			w := assert.NotWithin(t, syncBlockingTimeout, func(context.Context) { act(t) })
			holder.Get(t).Finish()
			assert.Within(t, syncTimeout, func(context.Context) { w.Wait() })

			assert.Contains(t, subject.Get(t).Logs.String(), key.Get(t))
			assert.Contains(t, subject.Get(t).Logs.String(), holder.Get(t).Name())
		})

		s.And("the other test holds a different key", func(s *testcase.Spec) {
			holderKey.Let(s, func(t *testcase.T) string {
				return key.Get(t) + "-other"
			})

			s.Then("it doesn't wait", func(t *testcase.T) {
				assert.Within(t, syncTimeout, func(context.Context) { act(t) })
			})
		})

		s.And("the other test's name is a prefix of the test's name, without being its parent", func(s *testcase.Spec) {
			subject.Let(s, func(t *testcase.T) *testcase.FakeTB {
				return newSyncTB(t, holder.Get(t).Name()+"2")
			})

			s.Then("it waits until the other test finishes", func(t *testcase.T) {
				w := assert.NotWithin(t, syncBlockingTimeout, func(context.Context) { act(t) })

				holder.Get(t).Finish()
				assert.Within(t, syncTimeout, func(context.Context) { w.Wait() })
			})
		})
	})

	s.When("the parent test holds the key", func(s *testcase.Spec) {
		parent := let.Var(s, func(t *testcase.T) *testcase.FakeTB {
			return newSyncTB(t, "TestParent")
		})
		subject.Let(s, func(t *testcase.T) *testcase.FakeTB {
			return newSyncTB(t, parent.Get(t).Name()+"/subtest")
		})
		s.Before(func(t *testcase.T) {
			testcase.Sync(parent.Get(t), key.Get(t))
		})

		s.Then("it doesn't wait", func(t *testcase.T) {
			assert.Within(t, syncTimeout, func(context.Context) { act(t) })
		})

		s.Then("other tests wait until the parent test finishes, not just the subtest", func(t *testcase.T) {
			act(t)
			other := newSyncTB(t, "TestOther")

			w := assert.NotWithin(t, syncBlockingTimeout, func(context.Context) {
				testcase.Sync(other, key.Get(t))
			})

			subject.Get(t).Finish()
			assert.NotWithin(t, syncBlockingTimeout, func(context.Context) { w.Wait() })

			parent.Get(t).Finish()
			assert.Within(t, syncTimeout, func(context.Context) { w.Wait() })
		})

		s.And("another subtest of the parent holds the key", func(s *testcase.Spec) {
			sibling := let.Var(s, func(t *testcase.T) *testcase.FakeTB {
				return newSyncTB(t, parent.Get(t).Name()+"/sibling")
			})
			s.Before(func(t *testcase.T) {
				testcase.Sync(sibling.Get(t), key.Get(t))
			})

			s.Then("it waits until the other subtest finishes", func(t *testcase.T) {
				w := assert.NotWithin(t, syncBlockingTimeout, func(context.Context) { act(t) })

				sibling.Get(t).Finish()
				assert.Within(t, syncTimeout, func(context.Context) { w.Wait() })
			})
		})
	})

	s.When("another test process holds the key", func(s *testcase.Spec) {
		process := let.Var(s, func(t *testcase.T) *syncHelperProcess {
			return startSyncHelperProcess(t, key.Get(t), "")
		}).EagerLoading(s)

		s.Then("it waits until the other process releases the key", func(t *testcase.T) {
			w := assert.NotWithin(t, syncBlockingTimeout, func(context.Context) { act(t) })

			process.Get(t).Release()
			assert.Within(t, syncTimeout, func(context.Context) { w.Wait() })
		})

		s.Then("the waiting is logged along with the details of the holder process", func(t *testcase.T) {
			w := assert.NotWithin(t, syncBlockingTimeout, func(context.Context) { act(t) })
			process.Get(t).Release()
			assert.Within(t, syncTimeout, func(context.Context) { w.Wait() })

			assert.Contains(t, subject.Get(t).Logs.String(), key.Get(t))
			assert.Contains(t, subject.Get(t).Logs.String(), "TestSync_helperProcess")
			assert.Contains(t, subject.Get(t).Logs.String(), fmt.Sprintf("pid: %d", process.Get(t).PID()))
		})

		s.And("the other process exits without releasing the key", func(s *testcase.Spec) {
			s.Before(func(t *testcase.T) {
				process.Get(t).Kill()
			})

			s.Then("it doesn't wait", func(t *testcase.T) {
				assert.Within(t, syncTimeout, func(context.Context) { act(t) })
			})
		})

		s.And("the other process belongs to a different project", func(s *testcase.Spec) {
			process.Let(s, func(t *testcase.T) *syncHelperProcess {
				dir := t.TempDir()
				gomod := []byte("module example.com/other\n")
				assert.Must(t).NoError(os.WriteFile(filepath.Join(dir, "go.mod"), gomod, 0o600))

				// the lock directory of the throwaway project is removed once the helper process exits
				var lockDir string
				t.Cleanup(func() {
					if lockDir != "" {
						_ = os.RemoveAll(lockDir)
					}
				})
				p := startSyncHelperProcess(t, key.Get(t), dir)
				lockDir = filepath.Dir(findSyncLockFile(t, key.Get(t)))
				return p
			})

			s.Then("it doesn't wait", func(t *testcase.T) {
				assert.Within(t, syncTimeout, func(context.Context) { act(t) })
			})
		})
	})
}

func newSyncTB(t *testcase.T, name string) *testcase.FakeTB {
	tb := &testcase.FakeTB{StubName: name}
	t.Cleanup(tb.Finish)
	return tb
}

// findSyncLockFile finds the lock file of the key in the lock directories of the projects,
// based on the holder details in the lock file.
func findSyncLockFile(t *testcase.T, key string) string {
	paths, err := filepath.Glob(filepath.Join(os.TempDir(), "testcase-*", "*.lock"))
	assert.Must(t).NoError(err)
	for _, path := range paths {
		data, err := os.ReadFile(path)
		if err == nil && bytes.Contains(data, []byte("key: "+key+"\n")) {
			return path
		}
	}
	t.Fatalf("the lock file of %q was not found", key)
	return ""
}

func fileMode(t *testcase.T, path string) fs.FileMode {
	info, err := os.Lstat(path)
	assert.Must(t).NoError(err)
	return info.Mode()
}

const (
	syncHelperProcessKeyEnv   = "SYNC_TEST_HELPER_PROCESS_KEY"
	syncHelperProcessHoldsKey = "SYNC_TEST_HELPER_PROCESS_HOLDS_KEY"
)

// TestSync_helperProcess is not a real test.
// TestSync runs it as a separate process to represent a concurrently running test process.
// It holds the key until its standard input is closed.
func TestSync_helperProcess(t *testing.T) {
	key, ok := os.LookupEnv(syncHelperProcessKeyEnv)
	if !ok {
		t.Skip("helper process of TestSync")
	}
	testcase.Sync(t, key)
	fmt.Println(syncHelperProcessHoldsKey)
	_, _ = io.Copy(io.Discard, os.Stdin)
}

type syncHelperProcess struct {
	cmd   *exec.Cmd
	stdin io.Closer
	done  chan struct{}
}

func startSyncHelperProcess(t *testcase.T, key, dir string) *syncHelperProcess {
	exe, err := os.Executable()
	assert.Must(t).NoError(err)

	var (
		cmd    = exec.Command(exe, "-test.run=^TestSync_helperProcess$")
		stdout = &syncHelperProcessOutput{holdsKey: make(chan struct{})}
		stderr bytes.Buffer
	)
	cmd.Dir = dir
	cmd.Env = append(os.Environ(),
		syncHelperProcessKeyEnv+"="+key,
		// a race enabled test binary waits a second before it exits, unless told otherwise
		"GORACE="+strings.TrimSpace(os.Getenv("GORACE")+" atexit_sleep_ms=0"),
	)
	cmd.Stdout = stdout
	cmd.Stderr = &stderr
	stdin, err := cmd.StdinPipe()
	assert.Must(t).NoError(err)
	assert.Must(t).NoError(cmd.Start())

	p := &syncHelperProcess{cmd: cmd, stdin: stdin, done: make(chan struct{})}
	go func() {
		defer close(p.done)
		_ = cmd.Wait()
	}()
	t.Cleanup(func() {
		p.Release()
		select {
		case <-p.done:
		case <-time.After(syncTimeout):
			p.Kill()
			<-p.done
		}
		if t.Failed() {
			t.Logf("helper process output:\n%s\n%s", stdout.String(), stderr.String())
		}
	})

	select {
	case <-stdout.holdsKey:
	case <-p.done:
		t.Fatalf("helper process exited without acquiring the key:\n%s\n%s", stdout.String(), stderr.String())
	case <-time.After(syncTimeout):
		t.Fatal("helper process didn't acquire the key in time")
	}
	return p
}

func (p *syncHelperProcess) PID() int { return p.cmd.Process.Pid }

// Release makes the helper process release the key and exit.
func (p *syncHelperProcess) Release() { _ = p.stdin.Close() }

// Kill makes the helper process exit without releasing the key.
func (p *syncHelperProcess) Kill() {
	_ = p.cmd.Process.Kill()
	<-p.done
}

type syncHelperProcessOutput struct {
	m        sync.Mutex
	buf      bytes.Buffer
	holdsKey chan struct{}
	closed   bool
}

func (o *syncHelperProcessOutput) Write(p []byte) (int, error) {
	o.m.Lock()
	defer o.m.Unlock()
	o.buf.Write(p)
	if !o.closed && bytes.Contains(o.buf.Bytes(), []byte(syncHelperProcessHoldsKey)) {
		o.closed = true
		close(o.holdsKey)
	}
	return len(p), nil
}

func (o *syncHelperProcessOutput) String() string {
	o.m.Lock()
	defer o.m.Unlock()
	return o.buf.String()
}

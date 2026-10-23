package testcase

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"go.llib.dev/testcase/internal/filelock"
	"go.llib.dev/testcase/random"
)

// Sync ensures that only a single test at a time uses the shared resource that the key identifies.
//
// Sync blocks until the key becomes available, then the test holds the key until it finishes.
// The exclusivity applies to every test within the current test process,
// and to the tests of other test processes of the same project on the same host.
// This is useful when the tests of several packages depend on the same external resource, like a database,
// since `go test ./...` runs the tests of each package as a separate process, at the same time.
//
// Different keys are independent, so they never block each other:
//
//	testcase.Sync(t, "foo")
//	testcase.Sync(t, "bar")
//	testcase.Sync(t, "baz")
//
// Calling Sync again with a key that the test already holds is a no-op.
// A subtest can Sync on a key that one of its parent tests holds,
// and while the subtest holds it, the other subtests of the parent wait for their turn.
//
// Call Sync before arranging the shared resource,
// so the resource's cleanups still run while the test holds the key.
//
// Test processes synchronize through lock files in the host's temporary directory,
// at "testcase-<project root path hash>/<key hash>.lock", where the project root is the closest directory with a go.mod file.
// The lock directory of the project belongs to the group of the project root directory,
// so the users who share the project through its group share the keys too.
// The operating system releases the lock of a test process when the process exits,
// so an interrupted test run can't leave a stale lock behind.
// On platforms without file locking support, Sync only synchronizes the tests within the current test process.
func Sync(tb testing.TB, key string) {
	tb.Helper()
	if key == "" {
		tb.Fatal("testcase.Sync: the key must not be empty")
	}
	release, err := syncs.acquire(tb, key)
	if err != nil {
		tb.Fatalf("testcase.Sync: unable to acquire %q: %v", key, err)
	}
	if release == nil { // the test already holds the key
		return
	}
	tb.Cleanup(release)
}

var syncs = newSyncRegistry()

// syncRegistry coordinates the keys between the tests of the current process,
// and acquires a key's host-wide lock when the first test of the process needs it.
type syncRegistry struct {
	m    sync.Mutex
	c    *sync.Cond
	keys map[string]*syncKey
}

type syncKey struct {
	// acquirer is the name of the test that is acquiring the host-wide lock of the key.
	acquirer string
	// holders is the chain of tests that hold the key.
	// Each holder is a subtest of the one before it, and the last holder is the one that uses the key.
	holders []string
	lock    *filelock.Lock
}

func newSyncRegistry() *syncRegistry {
	r := &syncRegistry{keys: make(map[string]*syncKey)}
	r.c = sync.NewCond(&r.m)
	return r
}

// acquire blocks until the test can use the key.
// It returns a nil release function when the test already holds the key.
func (r *syncRegistry) acquire(tb testing.TB, key string) (release func(), _ error) {
	tb.Helper()
	name := tb.Name()
	release, ok := r.claim(tb, key, name)
	if ok {
		return release, nil
	}
	lock, err := lockSyncFile(tb, key, name)
	r.m.Lock()
	defer r.m.Unlock()
	defer r.c.Broadcast()
	if err != nil {
		delete(r.keys, key)
		return nil, err
	}
	k := r.keys[key]
	k.acquirer = ""
	k.lock = lock
	k.holders = append(k.holders, name)
	return r.releaser(key, name), nil
}

// claim waits until the test can use the key within the current process.
// When nobody in the process holds the key, claim makes the test the key's acquirer,
// and reports false, as the test still has to acquire the host-wide lock of the key.
func (r *syncRegistry) claim(tb testing.TB, key, name string) (release func(), ok bool) {
	tb.Helper()
	r.m.Lock()
	defer r.m.Unlock()
	var waiting bool
	for {
		k, found := r.keys[key]
		if !found {
			r.keys[key] = &syncKey{acquirer: name}
			return nil, false
		}
		if k.acquirer == "" {
			for _, holder := range k.holders {
				if holder == name {
					return nil, true
				}
			}
			if user := k.holders[len(k.holders)-1]; strings.HasPrefix(name, user+"/") {
				k.holders = append(k.holders, name)
				return r.releaser(key, name), true
			}
		}
		if !waiting {
			waiting = true
			tb.Logf("testcase.Sync: waiting for %q, which is in use by %s", key, k.user())
		}
		r.c.Wait()
	}
}

func (r *syncRegistry) releaser(key, name string) func() {
	return func() {
		r.m.Lock()
		defer r.m.Unlock()
		defer r.c.Broadcast()
		k, ok := r.keys[key]
		if !ok {
			return
		}
		for i := len(k.holders) - 1; 0 <= i; i-- {
			if k.holders[i] == name {
				k.holders = append(k.holders[:i], k.holders[i+1:]...)
				break
			}
		}
		if 0 < len(k.holders) {
			return
		}
		delete(r.keys, key)
		k.lock.Unlock()
	}
}

func (k *syncKey) user() string {
	if k.acquirer != "" {
		return k.acquirer
	}
	return k.holders[len(k.holders)-1]
}

const syncMaxPollInterval = 100 * time.Millisecond

// lockSyncFile blocks until it acquires the host-wide lock of the key.
func lockSyncFile(tb testing.TB, key, name string) (*filelock.Lock, error) {
	tb.Helper()
	host, err := getSyncHost()
	if err != nil {
		return nil, err
	}
	var (
		path     = filepath.Join(host.dir, syncFileName(key))
		holder   = fmt.Sprintf("key: %s\ntest: %s\npid: %d\nprocess: %s\n", key, name, os.Getpid(), host.processID)
		interval = time.Millisecond
		waiting  bool
	)
	for {
		lock, ok, err := filelock.TryLock(path, []byte(holder))
		if err != nil || ok {
			return lock, err
		}
		if !waiting {
			waiting = true
			msg := fmt.Sprintf("testcase.Sync: waiting for %q, which is in use by another test process", key)
			if info, ok := filelock.ReadHolder(path); ok {
				msg += ":\n" + info
			}
			tb.Log(msg)
		}
		time.Sleep(interval)
		if interval *= 2; syncMaxPollInterval < interval {
			interval = syncMaxPollInterval
		}
	}
}

// syncFileName maps the key to a file name.
// The key itself is not suitable as a file name, since it might contain characters that a file name can't,
// and on case-insensitive file systems, keys that only differ in letter case would share the same file.
// The lock file's content still names the key, so a lock file can be traced back to its key.
func syncFileName(key string) string {
	sum := sha256.Sum256([]byte(key))
	return hex.EncodeToString(sum[:]) + ".lock"
}

type syncHost struct {
	// dir is where the lock files of the project are kept.
	dir string
	// processID identifies the current test process in the lock files.
	processID string
}

var syncHostInit = struct {
	once sync.Once
	host syncHost
	err  error
}{}

func getSyncHost() (syncHost, error) {
	syncHostInit.once.Do(func() {
		root, err := syncProjectRoot()
		if err != nil {
			syncHostInit.err = err
			return
		}
		dir, err := makeSyncDir(root)
		if err != nil {
			syncHostInit.err = err
			return
		}
		syncHostInit.host = syncHost{
			dir:       dir,
			processID: random.New(random.CryptoSeed{}).UUID(),
		}
	})
	return syncHostInit.host, syncHostInit.err
}

// makeSyncDir prepares the directory for the lock files of the project.
//
// The directory is only accessible to its owner and the group of the project root directory,
// so the users who share the project through its group can synchronize their test runs with each other.
// It is directly in the host's temporary directory, which usually lets only the owner of an entry remove or rename it,
// so other users can't tamper with it.
func makeSyncDir(root string) (string, error) {
	sum := sha256.Sum256([]byte(root))
	dir := filepath.Join(os.TempDir(), "testcase-"+hex.EncodeToString(sum[:]))
	if err := filelock.MkdirGroup(dir, root); err != nil {
		return "", err
	}
	return dir, nil
}

// syncProjectRoot returns the closest directory with a go.mod file, starting from the working directory.
// When there is no such directory, the working directory is the project root.
//
// The tests of each package run with their package directory as the working directory,
// so the closest go.mod file is the common point between the test processes of the same project.
func syncProjectRoot() (string, error) {
	wd, err := os.Getwd()
	if err != nil {
		return "", err
	}
	if path, err := filepath.EvalSymlinks(wd); err == nil {
		wd = path
	}
	for dir := wd; ; {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir, nil
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return wd, nil
		}
		dir = parent
	}
}

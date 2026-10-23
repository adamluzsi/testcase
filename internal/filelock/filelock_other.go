//go:build !darwin && !dragonfly && !freebsd && !linux && !netbsd && !openbsd && !windows

package filelock

// Lock is a no-op on platforms without file locking support.
type Lock struct{}

// TryLock always succeeds on platforms without file locking support,
// and it doesn't touch the file system.
func TryLock(string, []byte) (*Lock, bool, error) { return &Lock{}, true, nil }

func (*Lock) Unlock() {}

func ReadHolder(string) (string, bool) { return "", false }

func MkdirGroup(string, string) error { return nil }

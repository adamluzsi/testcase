package internal

import (
	"runtime"
	"strconv"
	"strings"
	"sync"
	"testing"
)

func GoID() int64 {
	buf := make([]byte, 64)
	n := runtime.Stack(buf, false)
	buf = buf[:n]
	idField := strings.Fields(strings.TrimPrefix(string(buf), "goroutine "))[0]
	id, _ := strconv.ParseInt(idField, 10, 64)
	return id
}

var gnp gNoParallel

type gNoParallel struct {
	m sync.Mutex

	goid *int64
}

func NoParallel(tb testing.TB, fatalmsg string) {
	goID := GoID()
	gotLock := gnp.m.TryLock()
	if !gotLock && gnp.goid != nil {
		if goID == *gnp.goid {
			return // we already own this goroutine, all good, nothing to do
		}
		tb.Fatal(fatalmsg)
	}
	gnp.goid = &goID
	tb.Cleanup(func() {
		gnp.goid = nil
		gnp.m.Unlock()
	})
	tb.Setenv("TESTING_FORBID_PARALLEL_EXECUTION", "-")
}

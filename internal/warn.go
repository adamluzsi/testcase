package internal

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"testing"
)

var warnOutput io.Writer = os.Stderr

func Warn(vs ...any) {
	out := append([]any{"[WARN]", "[TESTCASE]"}, vs...)
	fmt.Fprintln(warnOutput, out...)
}

func StubWarn(tb testing.TB) *bytes.Buffer {
	original := warnOutput
	tb.Cleanup(func() { warnOutput = original })
	var buf bytes.Buffer
	warnOutput = &buf
	return &buf
}

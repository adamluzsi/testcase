package pp

import (
	"fmt"
	"io"
	"os"
	"runtime"
	"strings"
	"sync"

	"go.llib.dev/testcase/internal/caller"
)

var defaultWriter io.Writer = os.Stderr
var l sync.Mutex

func PP(vs ...any) {
	l.Lock()
	defer l.Unlock()
	_, file, line, _ := runtime.Caller(1)
	_, _ = fmt.Fprintf(defaultWriter, "%s ", caller.AsLocation(true, file, line))
	_, _ = fpp(defaultWriter, vs...)
}

func FPP(w io.Writer, vs ...any) (int, error) {
	return fpp(w, vs...)
}

func fpp(w io.Writer, vs ...any) (int, error) {
	var (
		form string
		args []any
	)
	for _, v := range vs {
		form += "\t%s"
		args = append(args, Format(v))
	}
	form = strings.TrimPrefix(form, "\t")
	form += fmt.Sprintln()
	return fmt.Fprintf(w, form, args...)
}

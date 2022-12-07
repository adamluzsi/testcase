package pp

import (
	"fmt"
	"github.com/adamluzsi/testcase/internal/caller"
	"io"
	"os"
	"runtime"
	"strings"
)

var defaultWriter io.Writer = os.Stderr

func PP(vs ...any) {
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

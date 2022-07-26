package pp

import (
	"fmt"
	"io"
	"os"
)

var defaultWriter io.Writer = os.Stderr

func PP(vs ...any) {
	FPP(defaultWriter, vs...)
}

func FPP(w io.Writer, vs ...any) {
	var (
		form string
		args []any
	)
	for _, v := range vs {
		form += "%s\n"
		args = append(args, Format(v))
	}
	fmt.Fprintf(w, form, args...)
}

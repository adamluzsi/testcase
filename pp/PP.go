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
	for i, v := range vs {
		if i != 0 {
			form += "\t"
		}
		form += "%s"
		args = append(args, Format(v))
	}
	fmt.Fprintf(w, form+"\n", args...)
}

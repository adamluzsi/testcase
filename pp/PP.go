package pp

import (
	"fmt"
	"io"
	"os"
)

var defaultWriter io.Writer = os.Stderr

type labelled struct {
	Label string
	Value any
}

// L allows to label a value for pretty printing with PP.
func L(label string, v any) labelled {
	return labelled{
		Label: label,
		Value: v,
	}
}

func PP(vs ...any) {
	_, _ = FPP(defaultWriter, vs...)
}

func FPP(w io.Writer, vs ...any) (int, error) {
	var (
		form string
		args []any
	)
	for _, v := range vs {
		switch v := v.(type) {
		case labelled:
			form += "%s\t%s\n"
			args = append(args, v.Label, Format(v.Value))
		default:
			form += "%s\n"
			args = append(args, Format(v))
		}
	}
	return fmt.Fprintf(w, form, args...)
}

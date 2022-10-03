package pp

import (
	"bytes"
	"fmt"
	"strings"
	"text/tabwriter"
)

// CompactDiff format the values in pp.Format and print out the changes only in a line by line style.
func CompactDiff(v1, v2 any) string {
	return CompactDiffString(Format(v1), Format(v2))
}

// CompactDiffString compare strings line by line in a side-by-side style.
// The diff style is similar to GNU "diff -y".
func CompactDiffString(val, oth string) string {
	var (
		rows     []compactDiffRow
		valPos   int
		othPos   int
		valLines = toLines(val)
		othLines = toLines(oth)
	)
wrk:
	for {
		var (
			hasVal  bool
			hasOth  bool
			valLine string
			othLine string
		)
		if valPos < len(valLines) {
			hasVal = true
			valLine = valLines[valPos]
		}
		if othPos < len(othLines) {
			hasOth = true
			othLine = othLines[othPos]
		}
		if !hasVal && !hasOth {
			break wrk
		}
		// only "val" has more lines, "oth" is finished
		if hasVal && !hasOth {
			rows = append(rows, compactDiffRow{
				Sym:  "-",
				Line: valLine,
			})
			valPos++
			continue wrk
		}
		// only "oth" has more lines, "val" is finished
		if !hasVal && hasOth {
			rows = append(rows, compactDiffRow{
				Sym:  "+",
				Line: othLine,
			})
			othPos++
			continue wrk
		}

		if valLine == othLine {
			rows = append(rows, compactDiffRow{
				Sym:  "-",
				Line: valLine,
			}, compactDiffRow{
				Sym:  "+",
				Line: othLine,
			})
			valPos++
			othPos++
			continue wrk
		}
		///////////////////////////////////
		// not equals, both line present //
		///////////////////////////////////

		// othLine is part of "val",
		// flush out "val" lines until we reach the current "oth" line there
		if contains(valLines[valPos:], othLine) {
			rows = append(rows, compactDiffRow{
				Sym:  "-",
				Line: valLine,
			})
			valPos++
			continue wrk
		}

		// "val"'s line part of other eventually
		// flush out "oth" lines until we reach the current "val" line
		if contains(othLines[othPos:], valLine) {
			rows = append(rows, compactDiffRow{
				Sym:  "-",
				Line: othLine,
			})
			othPos++
			continue wrk
		}

		rows = append(rows, compactDiffRow{
			Sym:  " ",
			Line: valLine,
		})
		valPos++
		othPos++
		continue wrk
	}
	return toCompactDiffTable(rows)
}

type compactDiffRow struct {
	Sym  string
	Line string
}

func toCompactDiffTable(rows []compactDiffRow) string {
	escape := func(str string) string {
		return strings.ReplaceAll(str, "\t", "  ")
	}
	buf := &bytes.Buffer{}
	w := tabwriter.NewWriter(buf, 0, 0, 0, ' ', 0)
	for _, row := range rows {
		_, _ = fmt.Fprintf(w, "%s\t%s\n", row.Sym, escape(row.Line))
	}
	_ = w.Flush()
	return buf.String()
}

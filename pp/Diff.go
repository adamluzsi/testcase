package pp

import (
	"bufio"
	"bytes"
	"fmt"
	"strings"
	"text/tabwriter"
)

// Diff will pretty print two value and show side-by-side the difference between them.
func Diff[T any](v1, v2 T) {
	_, _ = defaultWriter.Write([]byte(DiffString(Format(v1), Format(v2))))
}

// DiffFormat format the values in pp.Format and compare the results line by line in a side-by-side style.
func DiffFormat[T any](v1, v2 T) string {
	return DiffString(Format(v1), Format(v2))
}

// DiffString compare strings line by line in a side-by-side style.
// The diff style is similar to GNU "diff -y".
func DiffString(val, oth string) string {
	var (
		rows     []diffTableRow
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
			rows = append(rows, diffTableRow{
				Left:      valLine,
				Right:     "",
				Separator: "<",
			})
			valPos++
			continue wrk
		}
		// only "oth" has more lines, "val" is finished
		if !hasVal && hasOth {
			rows = append(rows, diffTableRow{
				Left:      "",
				Right:     othLine,
				Separator: ">",
			})
			othPos++
			continue wrk
		}

		if valLine == othLine {
			rows = append(rows, diffTableRow{
				Left:      valLine,
				Right:     othLine,
				Separator: "",
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
			rows = append(rows, diffTableRow{
				Left:      valLine,
				Right:     "",
				Separator: "<",
			})
			valPos++
			continue wrk
		}

		// "val"'s line part of other eventually
		// flush out "oth" lines until we reach the current "val" line
		if contains(othLines[othPos:], valLine) {
			rows = append(rows, diffTableRow{
				Left:      "",
				Right:     othLine,
				Separator: ">",
			})
			othPos++
			continue wrk
		}

		rows = append(rows, diffTableRow{
			Left:      valLine,
			Right:     othLine,
			Separator: "|",
		})
		valPos++
		othPos++
		continue wrk
	}
	return toTable(rows)
}

type diffTableRow struct {
	Left      string
	Right     string
	Separator string
}

func contains(lines []string, str string) bool {
	for _, line := range lines {
		if line == str {
			return true
		}
	}
	return false
}

func toTable(rows []diffTableRow) string {
	var mLen int
	for _, row := range rows {
		if nLen := len(row.Left); mLen < nLen {
			mLen = nLen
		}
	}
	escape := func(str string) string {
		return strings.ReplaceAll(str, "\t", "  ")
	}
	padded := func(str string) string {
		//paddingLen := (mLen) / 2 / 2
		padding := strings.Repeat(" ", 2)
		return fmt.Sprintf("%s%s%s", padding, str, padding)
	}
	buf := &bytes.Buffer{}
	w := tabwriter.NewWriter(buf, 0, 0, 0, ' ', 0)
	for _, row := range rows {
		_, _ = fmt.Fprintf(w, "%s\t%s\t%s\n", escape(row.Left), padded(row.Separator), escape(row.Right))
	}
	_ = w.Flush()
	return buf.String()
}

func toLines(str string) []string {
	scanner := bufio.NewScanner(strings.NewReader(str))
	scanner.Split(bufio.ScanLines)
	var lines []string
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}
	_ = scanner.Err()
	return lines
}

package testcase

import (
	"bufio"
	"bytes"
	"fmt"
	"runtime"
	"strings"
)

func log(logger interface{ Logf(format string, args ...interface{}) }, args ...interface{}) {
	whiteSpace := strings.Repeat(` `, getWhitespaceCount())
	message := fmt.Sprintln(append([]interface{}{"\n"}, args...)...)
	logger.Logf("\r%s%s", whiteSpace, indentMessageLines(message))
}

// Aligns the provided message so that all lines after the first line start at the same location as the first line.
// Assumes that the first line starts at the correct location (after carriage return, tab, label, spacer and tab).
func indentMessageLines(message string) string {
	outBuf := new(bytes.Buffer)

	for i, scanner := 0, bufio.NewScanner(strings.NewReader(message)); scanner.Scan(); i++ {
		// no need to align first line because it starts at the correct location (after the label)
		if i != 0 {
			// append alignLen+1 spaces to align with "{{longestLabel}}:" before adding tab
			outBuf.WriteString("\n\t" + ` ` + "\t")
		}
		outBuf.WriteString(scanner.Text())
	}

	return outBuf.String()
}

// I'm unable to get the windows width during the test runtime,
// so I just make a guess that will work for 95% of the case.
func getWhitespaceCount() int {
	_, file, line, ok := runtime.Caller(1)
	if !ok {
		return 0
	}

	parts := strings.Split(file, "/")
	file = parts[len(parts)-1]
	length := 3 * len(fmt.Sprintf("%s:%d:        ", file, line))

	if length > 128 { // hard cap the allowed max erasing for now
		return 128
	}

	return length
}

package random

import (
	"bufio"
	"bytes"
	_ "embed"
	"fmt"
	"regexp"
	"sort"
)

var charset string

// blns is a Big List of Naughty strings.
// source: https://github.com/minimaxir/big-list-of-naughty-strings
//
//go:embed blns.txt
var blns []byte
var naughtyStrings []string

func init() {
	type CharsetRange struct {
		Start int
		End   int
	}
	for _, r := range []CharsetRange{
		{
			Start: 33,
			End:   126,
		},
		{
			Start: 161,
			End:   1159,
		},
		{
			Start: 1162,
			End:   1364,
		},
		{
			Start: 1567,
			End:   1610,
		},
		{
			Start: 1634,
			End:   1747,
		},
	} {
		for i := r.Start; i <= r.End; i++ {
			charset += string(rune(i))
		}
	}

	defer func() { blns = nil }() // free
	isComment := regexp.MustCompile(`^\s*#`)
	isBlank := regexp.MustCompile(`^\s*$`)
	scanner := bufio.NewScanner(bytes.NewReader(blns))
	scanner.Split(bufio.ScanLines)
	for scanner.Scan() {
		line := scanner.Text()
		if isComment.MatchString(line) {
			continue
		}
		if isBlank.MatchString(line) {
			continue
		}
		naughtyStrings = append(naughtyStrings, line)
	}
	if err := scanner.Err(); err != nil {
		fmt.Println("Error", "testcase/random", "blns:", err.Error())
	}
	sort.Strings(naughtyStrings)
}

package random

import (
	"bufio"
	"bytes"
	_ "embed"
	"fmt"
	"io/fs"
	"regexp"
	"sort"

	"github.com/adamluzsi/testcase/random/internal"
)

var charset string

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

	isComment := regexp.MustCompile(`^\s*#`)
	isBlank := regexp.MustCompile(`^\s*$`)
	if err := fs.WalkDir(internal.NaughtyStringsFS, "naughtystrings", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}

		data, err := internal.NaughtyStringsFS.ReadFile(path)
		if err != nil {
			return err
		}

		scanner := bufio.NewScanner(bytes.NewReader(data))
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
		return scanner.Err()
	}); err != nil {
		fmt.Println("Error", "testcase/random", "naughtystrings:", err.Error())
	}
	sort.Strings(naughtyStrings)
}

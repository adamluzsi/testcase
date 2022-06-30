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

var naughtyStrings []string

func init() {
	isComment := regexp.MustCompile(`^\s*#`)
	isBlank := regexp.MustCompile(`^\s*$`)
	if err := fs.WalkDir(internal.FixturesFS, "fixtures", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}

		data, err := internal.FixturesFS.ReadFile(path)
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
		fmt.Println("Error", "testcase/random", "fixtures:", err.Error())
	}
	sort.Strings(naughtyStrings)
}

package random

import (
	"bufio"
	"bytes"
	_ "embed"
	"fmt"
	"io/fs"
	"regexp"
	"sort"
	"strings"

	"github.com/adamluzsi/testcase/random/internal"
)

var fixtureStrings struct {
	naughty []string
	errors  []string
}

func init() {
	initNaughtyStrings()
	initErrorStrings()
}

func initErrorStrings() {
	errOut := func(err error) {
		fmt.Println("Error", "testcase/random", "fixtures:", err.Error())
	}

	data, err := internal.FixturesFS.ReadFile("fixtures/errors.txt")
	if err != nil {
		errOut(err)
		return
	}

	lines, err := extractLines(data)
	if err != nil {
		errOut(err)
		return
	}

	fixtureStrings.errors = append(fixtureStrings.errors, lines...)
	sort.Strings(fixtureStrings.errors)
}

func initNaughtyStrings() {
	if err := fs.WalkDir(internal.FixturesFS, "fixtures", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		if strings.Contains(path, "errors.txt") {
			return nil
		}

		data, err := internal.FixturesFS.ReadFile(path)
		if err != nil {
			return err
		}

		lines, err := extractLines(data)
		if err != nil {
			return err
		}
		fixtureStrings.naughty = append(fixtureStrings.naughty, lines...)
		return nil
	}); err != nil {
		fmt.Println("Error", "testcase/random", "fixtures:", err.Error())
	}
	sort.Strings(fixtureStrings.naughty)
}

var (
	isComment = regexp.MustCompile(`^\s*#`)
	isBlank   = regexp.MustCompile(`^\s*$`)
)

func extractLines(data []byte) ([]string, error) {
	var lines []string
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
		lines = append(lines, line)
	}
	return lines, scanner.Err()
}

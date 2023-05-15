package random

import (
	"bufio"
	"bytes"
	_ "embed"
	"fmt"
	"path"
	"regexp"
	"sort"

	"github.com/adamluzsi/testcase/random/internal"
)

var fixtureStrings struct {
	naughty      []string
	errors       []string
	emailDomains []string

	names struct {
		male   []string
		female []string
		last   []string
	}
}

func init() {
	fixtureStrings.naughty = getNaughtyStrings()
	fixtureStrings.errors = getLines("errors.txt")
	fixtureStrings.emailDomains = getLines("emaildomains.txt")
	fixtureStrings.names.last = getLines("contacts", "lastnames.txt")
	fixtureStrings.names.male = getLines("contacts", "malenames.txt")
	fixtureStrings.names.female = getLines("contacts", "femalenames.txt")
}

func getNaughtyStrings() []string {
	var ns []string
	ns = append(ns, getLines("blns.txt")...)
	ns = append(ns, getLines("nosql.txt")...)
	ns = append(ns, getLines("sql.txt")...)
	ns = append(ns, getLines("sqlerr.txt")...)
	sort.Strings(ns)
	return ns
}

func getLines(paths ...string) []string {
	filePath := path.Join("fixtures", path.Join(paths...))

	errOut := func(err error) {
		fmt.Println("Error", "testcase/random", "fixtures:", filePath, "err:", err.Error())
	}

	data, err := internal.FixturesFS.ReadFile(filePath)
	if err != nil {
		errOut(err)
		return nil
	}

	lines, err := extractLines(data)
	if err != nil {
		errOut(err)
		return nil
	}

	sort.Strings(lines)
	return lines
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

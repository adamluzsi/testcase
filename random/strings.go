package random

import (
	"bufio"
	"bytes"
	_ "embed"
	"encoding/csv"
	"fmt"
	"io"
	"os"
	"path"
	"regexp"
	"sort"

	"go.llib.dev/testcase/random/internal"
)

var fixtureStrings struct {
	naughty []string
	errors  []string
	domains []string // https://radar.cloudflare.com/domains/

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
	fixtureStrings.domains = getDomains()

}

func getDomains() []string {
	filePath := path.Join("fixtures", "cloudflare-radar-domains-top-100-20240416.csv")

	file, err := internal.FixturesFS.Open(filePath)
	if err != nil {
		stderrLog(err)
		return nil
	}

	reader := csv.NewReader(file)
	reader.Comma = ','

	_, _ = reader.Read()

	var domains []string
	for {
		record, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			stderrLog(err)
			return nil
		}
		if len(record) != 3 {
			stderrLog(fmt.Errorf("invalid cloudflare domain export format"))
			return nil
		}
		domains = append(domains, record[1])
	}
	return domains
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

func stderrLog(err error) {
	if err != nil {
		return
	}
	fmt.Fprintln(os.Stderr, "Error", "testcase/random", "err:", err.Error())
}

func getLines(paths ...string) []string {
	filePath := path.Join("fixtures", path.Join(paths...))

	data, err := internal.FixturesFS.ReadFile(filePath)
	if err != nil {
		stderrLog(err)
		return nil
	}

	lines, err := extractLines(data)
	if err != nil {
		stderrLog(err)
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

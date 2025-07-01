//go:generate mkdir -p naughtystrings
//go:generate wget "https://raw.githubusercontent.com/minimaxir/big-list-of-naughty-strings/master/blns.txt" -O "naughtystrings/blns.txt"
//go:generate wget "https://raw.githubusercontent.com/payloadbox/sql-injection-payload-list/master/Intruder/detect/Generic_SQLI.txt" -O "naughtystrings/sql.txt"
//go:generate wget "https://raw.githubusercontent.com/payloadbox/sql-injection-payload-list/master/Intruder/detect/Generic_ErrorBased.txt" -O "naughtystrings/sqlerr.txt"
//go:generate wget "https://raw.githubusercontent.com/payloadbox/sql-injection-payload-list/master/Intruder/detect/NoSQL/no-sql.txt" -O "naughtystrings/nosql.txt"
package fixture

// blns is a Big List of Naughty strings.
// source: https://github.com/minimaxir/big-list-of-naughty-strings

import (
	"bufio"
	"bytes"
	"embed"
	"encoding/csv"
	"fmt"
	"io"
	"os"
	"path"
	"regexp"
	"sort"
)

//go:embed assets/*
var Assets embed.FS

const fsDirName = "assets"

var Values struct {
	Naughty []string
	Errors  []string
	Domains []string // https://radar.cloudflare.com/domains/

	EmailDomains []string

	Names struct {
		Male   []string
		Female []string
		Last   []string
	}
}

func init() {
	Values.Naughty = getNaughtyStrings()
	Values.Errors = getLines("errors.txt")
	Values.EmailDomains = getLines("emaildomains.txt")
	Values.Names.Last = getLines("contacts", "lastnames.txt")
	Values.Names.Male = getLines("contacts", "malenames.txt")
	Values.Names.Female = getLines("contacts", "femalenames.txt")
	Values.Domains = getDomains()

}

func getDomains() []string {
	filePath := path.Join(fsDirName, "cloudflare-radar-domains-top-100-20240416.csv")

	file, err := Assets.Open(filePath)
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
	ns = append(ns, getLines("llm.txt")...)
	sort.Strings(ns)
	return ns
}

func stderrLog(err error) {
	if err == nil {
		return
	}
	fmt.Fprintln(os.Stderr, "Error", "testcase/random", "err:", err.Error())
}

func getLines(paths ...string) []string {
	filePath := path.Join(fsDirName, path.Join(paths...))

	data, err := Assets.ReadFile(filePath)
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

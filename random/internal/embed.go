//go:generate mkdir -p naughtystrings
//go:generate wget "https://raw.githubusercontent.com/minimaxir/big-list-of-naughty-strings/master/blns.txt" -O "naughtystrings/blns.txt"
//go:generate wget "https://raw.githubusercontent.com/payloadbox/sql-injection-payload-list/master/Intruder/detect/Generic_SQLI.txt" -O "naughtystrings/sql.txt"
//go:generate wget "https://raw.githubusercontent.com/payloadbox/sql-injection-payload-list/master/Intruder/detect/Generic_ErrorBased.txt" -O "naughtystrings/sqlerr.txt"
//go:generate wget "https://raw.githubusercontent.com/payloadbox/sql-injection-payload-list/master/Intruder/detect/NoSQL/no-sql.txt" -O "naughtystrings/nosql.txt"
package internal

// blns is a Big List of Naughty strings.
// source: https://github.com/minimaxir/big-list-of-naughty-strings

import "embed"

//go:embed fixtures/*
var FixturesFS embed.FS

package testcase

import (
	"os"
	"strings"
	"testing"
)

// WIP is a helper function that skips the test conditionally.
// The main purpose of the WIP test flag is to allow sharing partially finished code, which is still under development.
// The developer can flag the test they are actively working on with WIP+username, and these tests will only run for them.
//
// Global override:
//
//	If the WIP environment variable is set, the test always runs
//	(never skips), regardless of any other condition.
//	This is ideal for periodic CI jobs, to check if there are forgotten WIP tests.
//
// When no user name is provided:
//
//	The test is skipped completely for everyone.
//
// When user name(s) is/are provided:
//
//	The test is skipped unless the current user matches one of the provided
//	usernames. Matching is case-insensitive. The current user is resolved from
//	standard environment variables such as TESTCASE_USER, USER, USERNAME, and LOGNAME.
//
// Example:
//
//	// Skip unless WIP env var is set
//	testcase.WIP(t)
//
//	// Skip unless current user is "alice" or "bob" (case-insensitive)
//	testcase.WIP(t, "alice", "bob")
//
//	// WIP set → always run, regardless of users
//	export WIP=1
//	go test
func WIP(tb testing.TB, users ...string) {
	const skipMessage = "WIP"
	tb.Helper()

	// WIP env is a global override: always run if set
	if _, ok := os.LookupEnv("WIP"); ok {
		return
	}

	if _, ok := os.LookupEnv("NOWIP"); ok {
		tb.Skip(skipMessage)
	}

	if 0 < len(users) {
		for _, envKey := range usernameEnvVariables {
			if usr, ok := os.LookupEnv(envKey); ok && 0 < len(usr) {
				for _, exp := range users {
					if strings.EqualFold(usr, exp) {
						return
					}
				}
			}
		}
	}
	tb.Skip(skipMessage)
}

var usernameEnvVariables = []string{"TESTCASE_USER", "USER", "USERNAME", "LOGNAME", "UserName"}

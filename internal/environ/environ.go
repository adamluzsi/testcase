package environ

import (
	"os"
	"slices"
	"strings"

	"go.llib.dev/testcase/internal"
)

// KeySeed is the environment variable key that will be checked for a pseudo random seed,
// which will be used to randomize the order of executions between test cases.
const KeySeed = `TESTCASE_SEED`

// KeyOrdering is the environment variable key that will be checked for testCase determine
// what order of execution should be used between test cases in a testing group.
// The default sorting behavior is pseudo random based on an the seed.
//
// Mods:
// - defined: execute testCase in the order which they are being defined
// - random: pseudo random based ordering between tests.
const KeyOrdering = `TESTCASE_ORDERING`
const KeyOrdering2 = `TESTCASE_ORDER`

func OrderingKeys() []string {
	return []string{
		KeyOrdering,
		KeyOrdering2,
	}
}

const KeyDebug = "TESTCASE_DEBUG"

var acceptedKeys = []string{
	KeySeed,
	KeyOrdering,
	KeyOrdering2,
	KeyDebug,
}

func init() { CheckEnvKeys() }

func CheckEnvKeys() {
	var got bool
	for _, envPair := range os.Environ() {
		ekv := strings.SplitN(envPair, "=", 2) // best effort to split, but it might not be platform agnostic
		if len(ekv) != 2 {
			continue
		}
		key, _ := ekv[0], ekv[1]

		if !strings.HasPrefix(key, "TESTCASE_") {
			continue
		}

		if !slices.Contains(acceptedKeys, key) {
			got = true
			internal.Warn("unrecognised testcase variable:", key)
		}
	}
	if got {
		internal.Warn("check if you might have a typo.")
		internal.Warn("accepted environment variables:", strings.Join(acceptedKeys, ", "))
	}
}

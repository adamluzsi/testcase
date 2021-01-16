package testcase_test

import (
	"testing"

	"github.com/adamluzsi/testcase"
)

func ExampleSpec_Tag() {
	var t *testing.T
	s := testcase.NewSpec(t)

	s.Context(`E2E`, func(s *testcase.Spec) {
		// by tagging the spec spec, we can filter tests out later in our CI/CD pipeline.
		// A comma separated list can be set with TESTCASE_TAG_INCLUDE env variable to filter down to tests with certain tags.
		// And/Or a comma separated list can be provided with TESTCASE_TAG_EXCLUDE to exclude tests tagged with certain tags.
		s.Tag(`E2E`)

		s.Test(`some E2E testCase`, func(t *testcase.T) {
			// ...
		})
	})
}

// example usage:
// 	TESTCASE_TAG_INCLUDE='E2E' go testCase ./...
// 	TESTCASE_TAG_EXCLUDE='E2E' go testCase ./...
// 	TESTCASE_TAG_INCLUDE='E2E' TESTCASE_TAG_EXCLUDE='list,of,excluded,tags' go testCase ./...
//

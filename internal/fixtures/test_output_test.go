// Do not change this file, the test-output acceptance test depends on it.
package fixtures_test

import (
	"testing"

	"github.com/adamluzsi/testcase"
)

func TestFixtureOutput(t *testing.T) {
	if !testing.Verbose() {
		t.Skip()
	}
	s := testcase.NewSpec(t)
	s.Test(``, func(t *testcase.T) { t.Log(`foo`) })
	s.Test(``, func(t *testcase.T) { t.Log(`bar`) })
	s.Test(``, func(t *testcase.T) { t.Log(`baz`) })
}

package assert_test

import (
	"testing"

	"github.com/adamluzsi/testcase/assert"
)

func TestDiffFunc(t *testing.T) {
	diff := assert.DiffFunc(1, 2)
	if diff == "" {
		t.Fatalf("diff function returned empty value")
	}
	if diff != assert.DiffFunc(1, 2) {
		t.Fatalf("diff function is not deterministic")
	}
}

//go:build spike

package assert_test

import (
	"testing"

	"go.llib.dev/testcase/assert"
)

func Test_spike(t *testing.T) {
	type T struct {
		ID string
		V1 string
		V2 int
	}

	var vs []T

	assert.OneOf(t, vs, func(t testing.TB, got T) {
		assert.Equal(t, got.V1, "The Answer")
		assert.Equal(t, got.V2, 42)
	})
}

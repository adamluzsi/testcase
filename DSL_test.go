package testcase_test

import (
	"testing"

	"go.llib.dev/testcase"
	"go.llib.dev/testcase/assert"
	"go.llib.dev/testcase/internal"
	"go.llib.dev/testcase/internal/doubles"
)

func TestSpec_TODO(t *testing.T) {
	dtb := &doubles.TB{}
	internal.StubVerbose(t, true)

	todos := []string{"abc", "bcd", "cde"}
	s := testcase.NewSpec(dtb)
	for _, todo := range todos {
		s.TODO(todo)
	}
	s.Finish()
	dtb.Finish()

	assert.False(t, dtb.IsFailed)
	assert.False(t, dtb.IsSkipped)

	for _, todo := range todos {
		assert.OneOf(t, dtb.Tests, func(t assert.It, got *doubles.TB) {
			assert.Contain(t, got.Name(), "TODO: "+todo)
			assert.False(t, got.IsFailed)
			assert.True(t, got.IsSkipped)
		})
	}
}

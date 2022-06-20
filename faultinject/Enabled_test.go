package faultinject_test

import (
	"testing"

	"github.com/adamluzsi/testcase/assert"
	"github.com/adamluzsi/testcase/faultinject"
)

func Test_enabled(t *testing.T) {
	assert.True(t, faultinject.Enabled, "by default, fault injection should be enabled")
}

func TestForTest(t *testing.T) {
	og := faultinject.Enabled
	t.Cleanup(func() { faultinject.Enabled = og })

	t.Run("when .Enable is turned off globally", func(t *testing.T) {
		faultinject.Enabled = false
		t.Run("", func(t *testing.T) {
			faultinject.ForTest(t, true)
			assert.True(t, faultinject.Enabled)
		})
		assert.False(t, faultinject.Enabled, "after cleanup state is restored")
	})

	t.Run("when .Enable is turned on globally", func(t *testing.T) {
		faultinject.Enabled = true
		t.Run("", func(t *testing.T) {
			faultinject.ForTest(t, false)
			assert.False(t, faultinject.Enabled)
		})
		assert.True(t, faultinject.Enabled, "after cleanup state is restored")
	})
}

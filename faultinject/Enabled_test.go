package faultinject_test

import (
	"runtime"
	"testing"

	"github.com/adamluzsi/testcase"
	"github.com/adamluzsi/testcase/assert"
	"github.com/adamluzsi/testcase/faultinject"
)

func Test_enabled(t *testing.T) {
	assert.False(t, faultinject.Enabled(), "by default, fault injection should be disabled")
}

func TestEnabled_race(t *testing.T) {
	testcase.Race(func() {
		_ = faultinject.Enabled()
	}, func() {
		defer faultinject.Enable()()
	})
}

func TestEnable(t *testing.T) {
	CommonEnableTest(t, func(tb testing.TB) {
		tb.Cleanup(faultinject.Enable())
	})
}

func TestEnableForTest(t *testing.T) {
	CommonEnableTest(t, func(tb testing.TB) {
		faultinject.EnableForTest(tb)
	})
}

func CommonEnableTest(t *testing.T, act func(testing.TB)) {
	t.Run("enables fault injection during test than restore the og state", func(t *testing.T) {
		t.Run("", func(t *testing.T) {
			act(t)
			assert.True(t, faultinject.Enabled())
		})
		assert.False(t, faultinject.Enabled(), "after cleanup state is restored")
	})

	t.Run("when nested tests depend on Enabling, then only the last restores the original state", func(t *testing.T) {
		t.Run("", func(t *testing.T) {
			act(t)
			t.Run("", func(t *testing.T) {
				act(t)
			})
			assert.True(t, faultinject.Enabled(), "should be still enabled as the current testing scope is still active")
		})
		assert.False(t, faultinject.Enabled(), "after cleanup state is restored")
	})

	t.Run("when parallel tests depend on Enabling, then only the last finishing test will restore the original state", func(t *testing.T) {
		for i, m := 0, runtime.NumCPU()*4; i < m; i++ {
			t.Run("", func(t *testing.T) {
				t.Parallel()
				act(t)
				assert.Should(t).True(faultinject.Enabled())
			})
		}
	})
}

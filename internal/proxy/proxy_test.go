package proxy_test

import (
	"testing"
	"time"

	"go.llib.dev/testcase/assert"
	"go.llib.dev/testcase/internal/doubles"
	"go.llib.dev/testcase/internal/proxy"
)

func TestStubTimeNow(t *testing.T) {
	t.Run("default", func(t *testing.T) {
		now := time.Now()
		time.Sleep(time.Microsecond)
		assert.NotEqual(t, proxy.TimeNow(), now)
	})

	t.Run("stub", func(t *testing.T) {
		now := time.Now()

		var dtb doubles.TB
		proxy.StubTimeNow(&dtb, func() time.Time {
			return now
		})

		for i := 0; i < 42; i++ {
			assert.Equal(t, proxy.TimeNow(), now)
		}

		dtb.Finish()
		time.Sleep(time.Microsecond)
		assert.NotEqual(t, proxy.TimeNow(), now)
	})
}

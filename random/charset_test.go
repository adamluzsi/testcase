package random_test

import (
	"testing"

	"go.llib.dev/testcase/assert"
	"go.llib.dev/testcase/random"
)

func TestCharset(t *testing.T) {
	assert.NotEmpty(t, random.Charset())
	assert.NotEmpty(t, random.CharsetASCII())
	assert.NotEmpty(t, random.CharsetAlpha())
	assert.NotEmpty(t, random.CharsetDigit())
}

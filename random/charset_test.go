package random_test

import (
	"testing"

	"github.com/adamluzsi/testcase/assert"
	"github.com/adamluzsi/testcase/random"
)

func TestCharset(t *testing.T) {
	assert.NotEmpty(t, random.Charset())
	assert.NotEmpty(t, random.CharsetASCII())
	assert.NotEmpty(t, random.CharsetAlpha())
	assert.NotEmpty(t, random.CharsetDigit())
}

package pp_test

import (
	"strings"
	"testing"

	"github.com/adamluzsi/testcase/assert"
	"github.com/adamluzsi/testcase/pp"
)

const tmplFormatUnexportedFields = `
pp_test.X{
	a: 1,
	b: 2,
	c: 3,
	d: "4",
	e: map[int]int{
		5: 6,
	},
	f: []int{
		42,
	},
	g: make(chan string, 1),
}
`

func TestFormat_unexportedFields(t *testing.T) {
	type X struct {
		a int
		b uint
		c float32
		d string
		e map[int]int
		f []int
		g chan string
	}
	v := X{
		a: 1,
		b: 2,
		c: 3,
		d: "4",
		e: map[int]int{5: 6},
		f: []int{42},
		g: make(chan string, 1),
	}
	expected := strings.TrimSpace(tmplFormatUnexportedFields)
	assert.Equal(t, expected, pp.Format(v))
}

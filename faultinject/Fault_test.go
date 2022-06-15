package faultinject

import (
	"fmt"
	"testing"

	"github.com/adamluzsi/testcase/assert"
)

func TestNextFault(t *testing.T) {
	fs := make([]Fault, 0)

	_, ok := nextFault(&fs, func(fault Fault) bool {
		return true
	})
	assert.False(t, ok)

	ef1 := Fault{OnFunc: "1", Error: fmt.Errorf("0")}
	ef2 := Fault{OnFunc: "2", Error: fmt.Errorf("1")}
	ef3 := Fault{OnFunc: "2", Error: fmt.Errorf("2")}
	fs = append(fs, ef1, ef2, ef3)

	_, ok = nextFault(&fs, func(fault Fault) bool {
		return false
	})
	assert.False(t, ok)

	f, ok := nextFault(&fs, func(fault Fault) bool {
		return fault.OnFunc == "2"
	})
	assert.True(t, ok)
	assert.Equal(t, ef2, f)

	f, ok = nextFault(&fs, func(fault Fault) bool {
		return fault.OnFunc == "2"
	})
	assert.True(t, ok)
	assert.Equal(t, ef3, f)

	_, ok = nextFault(&fs, func(fault Fault) bool {
		return fault.OnFunc == "2"
	})
	assert.False(t, ok)

}

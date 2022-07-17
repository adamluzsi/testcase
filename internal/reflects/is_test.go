package reflects_test

import (
	"fmt"
	"testing"

	"github.com/adamluzsi/testcase/assert"
	"github.com/adamluzsi/testcase/internal/reflects"
)

func TestIsNil(t *testing.T) {
	type TestCase struct {
		V     func() any
		IsNil bool
	}

	type T struct{}

	for _, tc := range []TestCase{
		{
			V:     func() any { return nil },
			IsNil: true,
		},
		{
			V: func() any {
				var v *T
				return v
			},
			IsNil: true,
		},
		{
			V: func() any {
				return &T{}
			},
			IsNil: false,
		},
		{
			V: func() any {
				var v map[string]string
				return v
			},
			IsNil: true,
		},
		{
			V: func() any {
				return map[string]string{}
			},
			IsNil: false,
		},
		{
			V: func() any {
				var v []string
				return v
			},
			IsNil: true,
		},
		{
			V: func() any {
				return []string{}
			},
			IsNil: false,
		},
	} {
		tc := tc
		v := tc.V()
		t.Run(fmt.Sprintf("%T - %v", v, v), func(t *testing.T) {
			assert.Equal(t, tc.IsNil, reflects.IsNil(v))
		})
	}
}

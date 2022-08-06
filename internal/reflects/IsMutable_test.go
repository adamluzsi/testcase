package reflects_test

import (
	"fmt"
	"github.com/adamluzsi/testcase/assert"
	"github.com/adamluzsi/testcase/internal/reflects"
	"github.com/adamluzsi/testcase/random"
	"testing"
)

func TestIsMutable(t *testing.T) {
	rnd := random.New(random.CryptoSeed{})
	type TestCase struct {
		V  any
		Is bool
	}

	var nilPtr *int

	var av any
	ptrToNil := &av

	for _, tc := range []TestCase{
		{
			V:  nil,
			Is: false,
		},
		{
			V:  rnd.String(),
			Is: false,
		},
		{
			V:  rnd.Bool(),
			Is: false,
		},
		{
			V:  rnd.Int(),
			Is: false,
		},
		{
			V:  int8(rnd.Int()),
			Is: false,
		},
		{
			V:  int16(rnd.Int()),
			Is: false,
		},
		{
			V:  int32(rnd.Int()),
			Is: false,
		},
		{
			V:  int64(rnd.Int()),
			Is: false,
		},
		{
			V:  uint(rnd.Int()),
			Is: false,
		},
		{
			V:  uint8(rnd.Int()),
			Is: false,
		},
		{
			V:  uint16(rnd.Int()),
			Is: false,
		},
		{
			V:  uint32(rnd.Int()),
			Is: false,
		},
		{
			V:  uint64(rnd.Int()),
			Is: false,
		},
		{
			V:  rnd.Float32(),
			Is: false,
		},
		{
			V:  rnd.Float64(),
			Is: false,
		},
		{
			V:  complex64(0),
			Is: false,
		},
		{
			V:  complex128(0),
			Is: false,
		},
		{
			V:  struct{}{},
			Is: false,
		},
		{
			V:  &struct{}{},
			Is: true,
		},
		{
			V:  struct{ X *int }{},
			Is: true,
		},
		{
			V:  struct{ x *int }{},
			Is: true,
		},
		{
			V:  []struct{}{},
			Is: true,
		},
		{
			V:  map[int]struct{}{},
			Is: true,
		},
		{
			V:  make(chan int),
			Is: true,
		},
		{
			V:  nilPtr,
			Is: false,
		},
		{
			V:  ptrToNil,
			Is: true,
		},
	} {
		tc := tc
		t.Run(fmt.Sprintf("%T", tc.V), func(t *testing.T) {
			assert.Equal(t, tc.Is, reflects.IsMutable(tc.V))
		})
	}
}

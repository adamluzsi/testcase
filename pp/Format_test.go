package pp_test

import (
	"strings"
	"testing"

	"github.com/adamluzsi/testcase/assert"
	"github.com/adamluzsi/testcase/pp"
)

const FormatPartialOutput = `
pp_test.PrintStruct1{
	F1: "foo/bar/baz",
	F2: 42,
	F3: pp_test.PrintStruct2{
		F1: map[string]string{
			"baz": "qux",
			"foo": "bar",
		},
		F2: []string{
			"foo",
			"bar",
			"baz",
		},
		F3: []pp_test.PrintStruct3{
			pp_test.PrintStruct3{
				F1: (pp_test.SomeInterface)("Hello, world!"),
			},
		},
	},
}
`

func TestFormat_smoke(t *testing.T) {
	type SomeInterface interface{}
	type PrintStruct3 struct {
		F1 SomeInterface
	}
	type PrintStruct2 struct {
		F1 map[string]string
		F2 []string
		F3 []PrintStruct3
	}
	type PrintStruct1 struct {
		F1 string
		F2 int
		F3 PrintStruct2
	}
	v := PrintStruct1{
		F1: "foo/bar/baz",
		F2: 42,
		F3: PrintStruct2{
			F1: map[string]string{
				"foo": "bar",
				"baz": "qux",
			},
			F2: []string{"foo", "bar", "baz"},
			F3: []PrintStruct3{
				{F1: SomeInterface("Hello, world!")},
			},
		},
	}
	assert.Equal(t,
		strings.TrimSpace(FormatPartialOutput),
		pp.Format(v))
}

func TestFormat_nil(t *testing.T) {
	assert.Equal(t, "nil", pp.Format(nil))
}

func TestFormat_unexportedFields(t *testing.T) {
	type X struct {
		a int
		b uint
		c float32
		d string
		e map[int]int
	}
	v := X{
		a: 1,
		b: 2,
		c: 3,
		d: "4",
		e: map[int]int{5: 6},
	}
	expected := "pp_test.X{\n\ta: 1,\n\tb: 0x2,\n\tc: 3,\n\td: \"4\",\n\te: map[int]int{\n\t\t5: 6,\n\t},\n}"
	assert.Equal(t, expected, pp.Format(v))
}

func TestFormat_map(t *testing.T) {
	type TestCase struct {
		Desc string
		In   any
		Out  string
	}

	for _, tc := range []TestCase{
		{
			Desc: "map[string]...",
			In: map[string]int{
				"b": 42,
				"a": 42,
				"c": 42,
			},
			Out: "map[string]int{\n\t\"a\": 42,\n\t\"b\": 42,\n\t\"c\": 42,\n}",
		},
		{
			Desc: "map[int]...",
			In: map[int]int{
				2: 42,
				1: 42,
				3: 42,
			},
			Out: "map[int]int{\n\t1: 42,\n\t2: 42,\n\t3: 42,\n}",
		},
		{
			Desc: "map[int8]...",
			In: map[int8]int{
				2: 42,
				1: 42,
				3: 42,
			},
			Out: "map[int8]int{\n\t1: 42,\n\t2: 42,\n\t3: 42,\n}",
		},
		{
			Desc: "map[int8]...",
			In: map[int8]int{
				2: 42,
				1: 42,
				3: 42,
			},
			Out: "map[int8]int{\n\t1: 42,\n\t2: 42,\n\t3: 42,\n}",
		},
		{
			Desc: "map[int16]...",
			In: map[int16]int{
				2: 42,
				1: 42,
				3: 42,
			},
			Out: "map[int16]int{\n\t1: 42,\n\t2: 42,\n\t3: 42,\n}",
		},
		{
			Desc: "map[int32]...",
			In: map[int32]int{
				2: 42,
				1: 42,
				3: 42,
			},
			Out: "map[int32]int{\n\t1: 42,\n\t2: 42,\n\t3: 42,\n}",
		},
		{
			Desc: "map[int64]...",
			In: map[int64]int{
				2: 42,
				1: 42,
				3: 42,
			},
			Out: "map[int64]int{\n\t1: 42,\n\t2: 42,\n\t3: 42,\n}",
		},
		{
			Desc: "map[uint]...",
			In: map[uint]int{
				2: 42,
				1: 42,
				3: 42,
			},
			Out: "map[uint]int{\n\t0x1: 42,\n\t0x2: 42,\n\t0x3: 42,\n}",
		},
		{
			Desc: "map[uint8]...",
			In: map[uint8]int{
				2: 42,
				1: 42,
				3: 42,
			},
			Out: "map[uint8]int{\n\t0x1: 42,\n\t0x2: 42,\n\t0x3: 42,\n}",
		},
		{
			Desc: "map[uint16]...",
			In: map[uint16]int{
				2: 42,
				1: 42,
				3: 42,
			},
			Out: "map[uint16]int{\n\t0x1: 42,\n\t0x2: 42,\n\t0x3: 42,\n}",
		},
		{
			Desc: "map[uint32]...",
			In: map[uint32]int{
				2: 42,
				1: 42,
				3: 42,
			},
			Out: "map[uint32]int{\n\t0x1: 42,\n\t0x2: 42,\n\t0x3: 42,\n}",
		},
		{
			Desc: "map[uint64]...",
			In: map[uint64]int{
				2: 42,
				1: 42,
				3: 42,
			},
			Out: "map[uint64]int{\n\t0x1: 42,\n\t0x2: 42,\n\t0x3: 42,\n}",
		},
		{
			Desc: "map[float32]...",
			In: map[float32]int{
				2: 42,
				1: 42,
				3: 42,
			},
			Out: "map[float32]int{\n\t1: 42,\n\t2: 42,\n\t3: 42,\n}",
		},
		{
			Desc: "map[float64]...",
			In: map[float64]int{
				2: 42,
				1: 42,
				3: 42,
			},
			Out: "map[float64]int{\n\t1: 42,\n\t2: 42,\n\t3: 42,\n}",
		},
	} {
		t.Run(tc.Desc, func(t *testing.T) {
			assert.Equal(t,
				pp.Format(tc.In),
				strings.TrimSpace(tc.Out))
		})
	}
}

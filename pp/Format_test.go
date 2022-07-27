package pp_test

import (
	"fmt"
	"github.com/adamluzsi/testcase"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/adamluzsi/testcase/assert"
	"github.com/adamluzsi/testcase/pp"
)

func TestFormat(t *testing.T) {
	s := testcase.NewSpec(t)

	var v = testcase.Let[any](s, nil)
	act := func(t *testcase.T) string {
		return pp.Format(v.Get(t))
	}

	s.When("v is an uint", func(s *testcase.Spec) {
		v.Let(s, func(t *testcase.T) any {
			var i uint = uint(t.Random.Int())
			return i
		})

		s.Then("it will print out the uint value in decimal form", func(t *testcase.T) {
			t.Must.Equal(fmt.Sprintf("%d", v.Get(t)), act(t))
		})
	})

	s.When("v is an float", func(s *testcase.Spec) {
		v.Let(s, func(t *testcase.T) any {
			return 42.42
		})

		s.Then("it will print out the in a float format", func(t *testcase.T) {
			t.Must.Equal("42.42", act(t))
		})
	})

	s.When("v is a pointer", func(s *testcase.Spec) {
		s.Before(func(t *testcase.T) {
			t.Must.Equal(reflect.Pointer, reflect.ValueOf(v.Get(t)).Kind())
		})

		s.And("it is an uninitialized pointer value (e.g.: int)", func(s *testcase.Spec) {
			v.Let(s, func(t *testcase.T) any {
				var i *int
				return i
			})

			s.Then("it will be represented as nil", func(t *testcase.T) {
				t.Must.Equal("nil", act(t))
			})
		})

		s.And("it is an initialized pointer that points to a nil value (e.g.: interface{})", func(s *testcase.Spec) {
			v.Let(s, func(t *testcase.T) any {
				var interf interface{}
				return &interf
			})

			s.Then("it will return the address taking of the underling value that contains a nil value", func(t *testcase.T) {
				t.Must.Equal("&(interface {})(nil)", act(t))
			})
		})
	})

	s.When("v is a struct", func(s *testcase.Spec) {
		type T struct {
			Exported   int
			unexported int
		}
		v.Let(s, func(t *testcase.T) any {
			return T{
				Exported:   t.Random.Int(),
				unexported: t.Random.Int(),
			}
		})

		s.Then("it will print both the exported and the unexported fields", func(t *testcase.T) {
			t.Must.Equal(
				fmt.Sprintf("pp_test.T{\n\tExported: %d,\n\tunexported: %d,\n}",
					v.Get(t).(T).Exported,
					v.Get(t).(T).unexported),
				act(t))
		})
	})

	s.When("v is a time.Time", func(s *testcase.Spec) {
		v.Let(s, func(t *testcase.T) any {
			return time.Date(2022, time.July, 26, 17, 36, 19, 882377000, time.UTC)
		})

		s.Then("it will print out a time.Date() method constructor example", func(t *testcase.T) {
			expected := `time.Date(2022, time.July, 26, 17, 36, 19, 882377000, time.UTC)`
			t.Must.Equal(expected, act(t))
		})
	})

	s.When("v is a map", func(s *testcase.Spec) {
		v.Let(s, func(t *testcase.T) any {
			return map[int]struct{}{
				1: {},
				3: {},
				2: {},
			}
		})

		s.Then("it will print out a sorted map representation", func(t *testcase.T) {
			expected := "map[int]struct {}{\n\t1: struct {}{},\n\t2: struct {}{},\n\t3: struct {}{},\n}"
			t.Must.Equal(expected, act(t))
		})

		s.And("the values are nil", func(s *testcase.Spec) {
			v.Let(s, func(t *testcase.T) any {
				return map[int]*struct{}{
					4: nil,
					2: nil,
				}
			})

			s.Then("all the nil value is printed out", func(t *testcase.T) {
				expected := "map[int]*struct {}{\n\t2: nil,\n\t4: nil,\n}"
				t.Must.Equal(expected, act(t))
			})
		})
	})

}

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
	expected := "pp_test.X{\n\ta: 1,\n\tb: 2,\n\tc: 3,\n\td: \"4\",\n\te: map[int]int{\n\t\t5: 6,\n\t},\n}"
	assert.Equal(t, expected, pp.Format(v))
}

type ExampleFmtStringer []byte

func (e ExampleFmtStringer) String() string { return string(e) }

func TestFormat_fmtStringer(t *testing.T) {
	assert.Equal(t, `"foo/bar/baz"`, pp.Format(ExampleFmtStringer("foo/bar/baz")))
}

func TestFormat_timeTime(t *testing.T) {
	tm := time.Date(2022, time.July, 26, 17, 36, 19, 882377000, time.UTC)
	assert.Equal(t, `time.Date(2022, time.July, 26, 17, 36, 19, 882377000, time.UTC)`, pp.Format(tm))
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
			Out: "map[uint]int{\n\t1: 42,\n\t2: 42,\n\t3: 42,\n}",
		},
		{
			Desc: "map[uint8]...",
			In: map[uint8]int{
				2: 42,
				1: 42,
				3: 42,
			},
			Out: "map[uint8]int{\n\t1: 42,\n\t2: 42,\n\t3: 42,\n}",
		},
		{
			Desc: "map[uint16]...",
			In: map[uint16]int{
				2: 42,
				1: 42,
				3: 42,
			},
			Out: "map[uint16]int{\n\t1: 42,\n\t2: 42,\n\t3: 42,\n}",
		},
		{
			Desc: "map[uint32]...",
			In: map[uint32]int{
				2: 42,
				1: 42,
				3: 42,
			},
			Out: "map[uint32]int{\n\t1: 42,\n\t2: 42,\n\t3: 42,\n}",
		},
		{
			Desc: "map[uint64]...",
			In: map[uint64]int{
				2: 42,
				1: 42,
				3: 42,
			},
			Out: "map[uint64]int{\n\t1: 42,\n\t2: 42,\n\t3: 42,\n}",
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
				strings.TrimSpace(tc.Out),
				pp.Format(tc.In))
		})
	}
}

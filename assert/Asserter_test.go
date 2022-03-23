package assert_test

import (
	"errors"
	"fmt"
	"reflect"
	"strings"
	"testing"

	"github.com/adamluzsi/testcase/assert"
	"github.com/adamluzsi/testcase/fixtures"
	"github.com/adamluzsi/testcase/internal"
)

func TestMust(t *testing.T) {
	h := assert.Must(t)
	var failedNow bool
	stub := &internal.StubTB{StubFailNow: func() { failedNow = true }}
	a := assert.Must(stub)
	a.True(false) // fail it
	h.True(failedNow)
	h.True(stub.IsFailed)
}

func TestShould(t *testing.T) {
	h := assert.Must(t)
	var failedNow bool
	stub := &internal.StubTB{StubFailNow: func() { failedNow = true }}
	a := assert.Should(stub)
	a.True(false) // fail it
	h.True(!failedNow)
	h.True(stub.IsFailed)
}

func asserter(failFn func(args ...interface{})) assert.Asserter {
	return assert.Asserter{TB: &internal.StubTB{}, Fn: failFn}
}

func Equal(tb testing.TB, a, b interface{}) {
	tb.Helper()
	if !reflect.DeepEqual(a, b) {
		tb.Fatalf("A and B not equal: %#v <=> %#v", a, b)
	}
}

func AssertFailFnArgs(tb testing.TB, expected, output []interface{}) {
	tb.Helper()
	tb.Cleanup(func() {
		if tb.Failed() {
			tb.Log("output:")
			tb.Logf("\t%#v", output)
			tb.Logf("\n\t%s", output)
			tb.Log()
			tb.Log("expected:")
			tb.Logf("\t%#v", expected)
			tb.Logf("\n\t%s", expected)
		}
	})
	join := func(vs []interface{}) string {
		return strings.TrimSpace(fmt.Sprintln(vs...))
	}
	if strings.Contains(join(output), join(expected)) {
		return
	}
	tb.Fatalf("expected msg not found\noutput: %#v\nexpected: %#v", output, expected)
}

func TestAsserter_True(t *testing.T) {
	t.Run(`when true passed`, func(t *testing.T) {
		var failed bool
		subject := asserter(func(args ...interface{}) { failed = true })
		subject.True(true)
		Equal(t, failed, false)
	})
	t.Run(`when false passed`, func(t *testing.T) {
		var failed bool
		var actualMsg []interface{}
		subject := asserter(func(args ...interface{}) {
			failed = true
			actualMsg = args
		})
		expectedMsg := []interface{}{"foo", "bar", "baz"}
		subject.True(false, expectedMsg...)
		Equal(t, failed, true)
		AssertFailFnArgs(t, expectedMsg, actualMsg)
	})
}

func TestAsserter_False(t *testing.T) {
	t.Run(`when true passed`, func(t *testing.T) {
		var failed bool
		var actualMsg []interface{}
		subject := asserter(func(args ...interface{}) {
			failed = true
			actualMsg = args
		})
		expectedMsg := []interface{}{"hello", "world", 42}
		subject.False(true, expectedMsg...)
		Equal(t, failed, true)
		AssertFailFnArgs(t, expectedMsg, actualMsg)
	})
	t.Run(`when false passed`, func(t *testing.T) {
		var failed bool
		subject := asserter(func(args ...interface{}) { failed = true })
		subject.False(false)
		Equal(t, failed, false)

	})
}

func TestAsserter_Nil(t *testing.T) {
	t.Run(`when nil passed, then it is accepted`, func(t *testing.T) {
		var failed bool
		subject := asserter(func(args ...interface{}) { failed = true })
		subject.Nil(nil)
		Equal(t, failed, false)
	})
	t.Run(`when pointer with nil value passed, then it is accepted as nil`, func(t *testing.T) {
		var failed bool
		subject := asserter(func(args ...interface{}) { failed = true })
		subject.Nil((func())(nil))
		Equal(t, failed, false)
	})
	t.Run(`when non nil value is passed`, func(t *testing.T) {
		var failed bool
		var actualMsg []interface{}
		subject := asserter(func(args ...interface{}) {
			failed = true
			actualMsg = args
		})
		expectedMsg := []interface{}{"foo", "bar", "baz"}
		subject.Nil(errors.New("not nil"), expectedMsg...)
		Equal(t, failed, true)
		AssertFailFnArgs(t, expectedMsg, actualMsg)
	})
	t.Run("when non nil zero value is passed", func(t *testing.T) {
		var failed bool
		var actualMsg []interface{}
		subject := asserter(func(args ...interface{}) {
			failed = true
			actualMsg = args
		})
		expectedMsg := []interface{}{"foo", "bar", "baz"}
		subject.Nil("", expectedMsg...) // zero string value
		Equal(t, failed, true)
		AssertFailFnArgs(t, expectedMsg, actualMsg)
	})
}

func TestAsserter_NotNil(t *testing.T) {
	t.Run(`when nil passed`, func(t *testing.T) {
		var out []interface{}
		subject := asserter(func(args ...interface{}) { out = args })
		msg := []interface{}{"foo", "bar", "baz"}
		subject.NotNil(nil, msg...)
		AssertFailFnArgs(t, msg, out)
	})
	t.Run(`when pointer with nil value passed, then it is refused as nil`, func(t *testing.T) {
		var failed bool
		subject := asserter(func(args ...interface{}) { failed = true })
		subject.NotNil((func())(nil))
		Equal(t, failed, true)
	})
	t.Run(`when non nil value is passed`, func(t *testing.T) {
		var failed bool
		subject := asserter(func(args ...interface{}) { failed = true })
		subject.NotNil(errors.New("not nil"), "foo", "bar", "baz")
		Equal(t, failed, false)
	})
	t.Run("when non nil zero value is passed", func(t *testing.T) {
		var failed bool
		subject := asserter(func(args ...interface{}) { failed = true })
		subject.NotNil("", "foo", "bar", "baz")
		Equal(t, failed, false)
	})
}

func TestAsserter_Panics(t *testing.T) {
	t.Run(`when no panic, fails`, func(t *testing.T) {
		var failed bool
		var out []interface{}
		subject := asserter(func(args ...interface{}) { failed = true; out = args })
		subject.Panic(func() { /* nothing */ }, "boom!")
		Equal(t, failed, true)
		AssertFailFnArgs(t, []interface{}{"boom!"}, out)
	})
	t.Run(`when panic with nil value, pass`, func(t *testing.T) {
		var failed bool
		subject := asserter(func(args ...interface{}) { failed = true })
		subject.Panic(func() { panic(nil) }, "boom!")
		Equal(t, failed, false)
	})
	t.Run(`when panic with something, pass`, func(t *testing.T) {
		var failed bool
		subject := asserter(func(args ...interface{}) { failed = true })
		subject.Panic(func() { panic("something") }, "boom!")
		Equal(t, failed, false)
	})
}

func TestAsserter_NotPanics(t *testing.T) {
	t.Run(`when no panic, pass`, func(t *testing.T) {
		var failed bool
		subject := asserter(func(args ...interface{}) { failed = true })
		subject.NotPanic(func() { /* nothing */ }, "boom!")
		Equal(t, failed, false)
	})
	t.Run(`when panic with nil value, fail`, func(t *testing.T) {
		var failed bool
		var out []interface{}
		subject := asserter(func(args ...interface{}) { failed = true; out = args })
		subject.NotPanic(func() { panic(nil) }, "boom!")
		Equal(t, failed, true)
		AssertFailFnArgs(t, []interface{}{"boom!"}, out)
	})
	t.Run(`when panic with something, fail`, func(t *testing.T) {
		var failed bool
		var out []interface{}
		subject := asserter(func(args ...interface{}) { failed = true; out = args })
		subject.NotPanic(func() { panic("something") }, "boom!")
		Equal(t, failed, true)
		AssertFailFnArgs(t, []interface{}{"boom!"}, out)
		AssertFailFnArgs(t, []interface{}{"something"}, out)
	})
}

func TestAsserter_Equal(t *testing.T) {
	type TestCase struct {
		Desc     string
		Expected interface{}
		Actual   interface{}
		IsFailed bool
	}
	type E struct{ V int }

	//fn1 := func() {}
	//fn2 := func() {}

	for _, tc := range []TestCase{
		{
			Desc:     "when two basic type provided - int - equals",
			Expected: 42,
			Actual:   42,
			IsFailed: false,
		},
		{
			Desc:     "when two basic type provided - int - not equal",
			Expected: 42,
			Actual:   24,
			IsFailed: true,
		},
		{
			Desc:     "when two basic type provided - string - equals",
			Expected: "42",
			Actual:   "42",
			IsFailed: false,
		},
		{
			Desc:     "when two basic type provided - string - not equal",
			Expected: "42",
			Actual:   "24",
			IsFailed: true,
		},
		{
			Desc:     "when struct is provided - equals",
			Expected: E{V: 42},
			Actual:   E{V: 42},
			IsFailed: false,
		},
		{
			Desc:     "when struct is provided - not equal",
			Expected: E{V: 42},
			Actual:   E{V: 24},
			IsFailed: true,
		},
		{
			Desc:     "when struct ptr is provided - equals",
			Expected: &E{V: 42},
			Actual:   &E{V: 42},
			IsFailed: false,
		},
		{
			Desc:     "when struct ptr is provided - not equal",
			Expected: &E{V: 42},
			Actual:   &E{V: 24},
			IsFailed: true,
		},
		{
			Desc:     "when byte slice is provided - equals",
			Expected: []byte("foo"),
			Actual:   []byte("foo"),
			IsFailed: false,
		},
		{
			Desc:     "when byte slice is provided - not equal",
			Expected: []byte("foo"),
			Actual:   []byte("bar"),
			IsFailed: true,
		},
		{
			Desc:     "when byte slice is provided - not equal - expected populated, actual nil",
			Expected: []byte("foo"),
			Actual:   nil,
			IsFailed: true,
		},
		{
			Desc:     "when byte slice is provided - not equal - expected nil, actual populated",
			Expected: nil,
			Actual:   []byte("foo"),
			IsFailed: true,
		},
		//{
		//	Desc:     "when equal function provided",
		//	Expected: fn1,
		//	Actual:   fn1,
		//	IsFailed: false,
		//},
		//{
		//	Desc:     "when not equal functions provided",
		//	Expected: fn1,
		//	Actual:   fn2,
		//	IsFailed: true,
		//},
	} {
		tc := tc
		t.Run(tc.Desc, func(t *testing.T) {
			expectedMsg := []interface{}{fixtures.Random.StringN(3), fixtures.Random.StringN(3)}
			var actualMsg []interface{}
			var failed bool
			subject := asserter(func(args ...interface{}) {
				failed = true
				actualMsg = args
			})

			subject.Equal(tc.Expected, tc.Actual, expectedMsg...)
			if tc.IsFailed && actualMsg != nil {
				t.Log(actualMsg...)
			}

			Equal(t, failed, tc.IsFailed)
			if !tc.IsFailed {
				return
			}

			AssertFailFnArgs(t, expectedMsg, actualMsg)
		})
	}
}

func TestAsserter_NotEqual(t *testing.T) {
	type TestCase struct {
		Desc     string
		Expected interface{}
		Actual   interface{}
		IsFailed bool
	}
	type E struct{ V int }

	for _, tc := range []TestCase{
		{
			Desc:     "when two basic type provided - int - equals",
			Expected: 42,
			Actual:   42,
			IsFailed: true,
		},
		{
			Desc:     "when two basic type provided - int - not equal",
			Expected: 42,
			Actual:   24,
			IsFailed: false,
		},
		{
			Desc:     "when two basic type provided - string - equals",
			Expected: "42",
			Actual:   "42",
			IsFailed: true,
		},
		{
			Desc:     "when two basic type provided - string - not equal",
			Expected: "42",
			Actual:   "24",
			IsFailed: false,
		},
		{
			Desc:     "when struct is provided - equals",
			Expected: E{V: 42},
			Actual:   E{V: 42},
			IsFailed: true,
		},
		{
			Desc:     "when struct is provided - not equal",
			Expected: E{V: 42},
			Actual:   E{V: 24},
			IsFailed: false,
		},
		{
			Desc:     "when struct ptr is provided - equals",
			Expected: &E{V: 42},
			Actual:   &E{V: 42},
			IsFailed: true,
		},
		{
			Desc:     "when struct ptr is provided - not equal",
			Expected: &E{V: 42},
			Actual:   &E{V: 24},
			IsFailed: false,
		},
		{
			Desc:     "when byte slice is provided - equals",
			Expected: []byte("foo"),
			Actual:   []byte("foo"),
			IsFailed: true,
		},
		{
			Desc:     "when byte slice is provided - not equal",
			Expected: []byte("foo"),
			Actual:   []byte("bar"),
			IsFailed: false,
		},
		{
			Desc:     "when byte slice is provided - not equal - expected populated, actual nil",
			Expected: []byte("foo"),
			Actual:   nil,
			IsFailed: false,
		},
		{
			Desc:     "when byte slice is provided - not equal - expected nil, actual populated",
			Expected: nil,
			Actual:   []byte("foo"),
			IsFailed: false,
		},
	} {
		tc := tc
		t.Run(tc.Desc, func(t *testing.T) {
			expectedMsg := []interface{}{fixtures.Random.StringN(3), fixtures.Random.StringN(3)}
			var actualMsg []interface{}
			var failed bool
			subject := asserter(func(args ...interface{}) {
				failed = true
				actualMsg = args
			})

			subject.NotEqual(tc.Expected, tc.Actual, expectedMsg...)
			if tc.IsFailed && actualMsg != nil {
				t.Log(actualMsg...)
			}

			Equal(t, failed, tc.IsFailed)
			if !tc.IsFailed {
				return
			}

			AssertFailFnArgs(t, expectedMsg, actualMsg)
		})
	}
}

func AssertContainsWith(tb testing.TB, isFailed bool, contains func(a assert.Asserter, msg []interface{})) {
	tb.Helper()

	expectedMsg := []interface{}{fixtures.Random.StringN(3), fixtures.Random.StringN(3)}
	var actualMsg []interface{}
	var failed bool
	subject := asserter(func(args ...interface{}) {
		failed = true
		actualMsg = args
	})

	contains(subject, expectedMsg)
	if isFailed && actualMsg != nil {
		tb.Log(actualMsg...)
	}

	Equal(tb, failed, isFailed)
	if !isFailed {
		return
	}

	// at this point slice contains behavior is confirmed
	AssertFailFnArgs(tb, expectedMsg, actualMsg)
}

func AssertContainsTestCase(src, has interface{}, isFailed bool) func(*testing.T) {
	return func(t *testing.T) {
		t.Helper()

		AssertContainsWith(t, isFailed, func(a assert.Asserter, msg []interface{}) {
			a.Contain(src, has, msg...)
		})
	}
}

func AssertContainExactlyTestCase(src, oth interface{}, isFailed bool) func(*testing.T) {
	return func(t *testing.T) {
		t.Helper()

		AssertContainsWith(t, isFailed, func(a assert.Asserter, msg []interface{}) {
			a.ContainExactly(src, oth, msg...)
		})
	}
}

func TestAsserter_Contain_invalid(t *testing.T) {
	t.Run(`when source is invalid`, func(t *testing.T) {
		var out []interface{}
		asserter(func(args ...interface{}) { out = args }).Contain(nil, []int{42})
		AssertFailFnArgs(t, []interface{}{"invalid source value"}, out)
	})
	t.Run(`when "has" is invalid`, func(t *testing.T) {
		var out []interface{}
		asserter(func(args ...interface{}) { out = args }).Contain([]int{42}, nil)
		AssertFailFnArgs(t, []interface{}{`invalid "has" value`}, out)
	})
}

func TestAsserter_Contain_typeMismatch(t *testing.T) {
	assert.Must(t).Panic(func() {
		asserter(func(args ...interface{}) {}).Contain([]int{42}, []string{"42"})
	}, "will panic on type mismatch")
	assert.Must(t).Panic(func() {
		asserter(func(args ...interface{}) {}).Contain([]int{42}, "42")
	}, "will panic on type mismatch")
}

func TestAsserter_Contain_sliceHasSubSlice(t *testing.T) {
	type TestCase struct {
		Desc     string
		Slice    interface{}
		Contains interface{}
		IsFailed bool
	}

	for _, tc := range []TestCase{
		{
			Desc:     "[]int: when equals",
			Slice:    []int{42, 24},
			Contains: []int{42, 24},
			IsFailed: false,
		},
		{
			Desc:     "[]int: when doesn't have all the elements",
			Slice:    []int{42, 24},
			Contains: []int{42, 24, 42},
			IsFailed: true,
		},
		{
			Desc:     "[]int: when fully includes in the beginning",
			Slice:    []int{42, 24, 4, 2, 2, 4},
			Contains: []int{42, 24},
			IsFailed: false,
		},
		{
			Desc:     "[]int: when fully includes in the end",
			Slice:    []int{4, 2, 2, 4, 42, 24},
			Contains: []int{42, 24},
			IsFailed: false,
		},
		{
			Desc:     "[]int: when fully includes in the middle",
			Slice:    []int{4, 2, 42, 24, 2, 4},
			Contains: []int{42, 24},
			IsFailed: false,
		},
		{
			Desc:     "[]string: when equals",
			Slice:    []string{"42", "24"},
			Contains: []string{"42", "24"},
			IsFailed: false,
		},
		{
			Desc:     "[]string: when slice has less element that the sub slice",
			Slice:    []string{"42", "24"},
			Contains: []string{"42", "24", "42"},
			IsFailed: true,
		},
		{
			Desc:     "[]string: when doesn't have fully matching elements",
			Slice:    []string{"42", "42"},
			Contains: []string{"42", "41"},
			IsFailed: true,
		},
		{
			Desc:     "[]string: when fully includes in the beginning",
			Slice:    []string{"42", "24", "4", "2", "2", "4"},
			Contains: []string{"42", "24"},
			IsFailed: false,
		},
		{
			Desc:     "[]string: when fully includes in the end",
			Slice:    []string{"4", "2", "2", "4", "42", "24"},
			Contains: []string{"42", "24"},
			IsFailed: false,
		},
		{
			Desc:     "[]string: when fully includes in the middle",
			Slice:    []string{"4", "2", "42", "24", "2", "4"},
			Contains: []string{"42", "24"},
			IsFailed: false,
		},
		{
			Desc:     "[]string: when fully includes in the middle",
			Slice:    []string{"4", "2", "42", "24", "2", "4"},
			Contains: []string{"42", "24"},
			IsFailed: false,
		},
	} {
		t.Run(tc.Desc, AssertContainsTestCase(tc.Slice, tc.Contains, tc.IsFailed))
	}
}

func TestAsserter_Contain_map(t *testing.T) {
	type TestCase struct {
		Desc     string
		Map      interface{}
		Has      interface{}
		IsFailed bool
	}

	for _, tc := range []TestCase{
		{
			Desc:     "when equals",
			Map:      map[int]int{42: 42, 24: 24},
			Has:      map[int]int{42: 42, 24: 24},
			IsFailed: false,
		},
		{
			Desc:     "when doesn't have all the elements",
			Map:      map[int]int{42: 42, 24: 24},
			Has:      map[int]int{42: 42, 24: 24, 12: 12},
			IsFailed: true,
		},
		{
			Desc:     "when map contains sub map",
			Map:      map[int]int{42: 42, 24: 24},
			Has:      map[int]int{42: 42},
			IsFailed: false,
		},
		{
			Desc:     "when map contains sub map keys but with different value",
			Map:      map[int]int{42: 24, 24: 42},
			Has:      map[int]int{42: 42},
			IsFailed: true,
		},
		{
			Desc:     "when map contains sub map keys, and values are nil",
			Map:      map[int]*int{42: nil, 24: nil},
			Has:      map[int]*int{42: nil},
			IsFailed: false,
		},
		{
			Desc:     "when map contains sub map keys, and the key is nil",
			Map:      map[*int]int{nil: 42},
			Has:      map[*int]int{nil: 42},
			IsFailed: false,
		},
	} {
		t.Run(tc.Desc, AssertContainsTestCase(tc.Map, tc.Has, tc.IsFailed))
	}
}

func TestAsserter_Contain_sliceHasElement(t *testing.T) {
	type TestCase struct {
		Desc     string
		Slice    interface{}
		Contains interface{}
		IsFailed bool
	}

	for _, tc := range []TestCase{
		{
			Desc:     "int: when doesn't have the element",
			Slice:    []int{42, 24},
			Contains: 12,
			IsFailed: true,
		},
		{
			Desc:     "int: when has the value in the beginning",
			Slice:    []int{42, 24, 4, 2, 2, 4},
			Contains: 42,
			IsFailed: false,
		},
		{
			Desc:     "int: when has the value includes in the end",
			Slice:    []int{4, 2, 2, 4, 42, 24},
			Contains: 42,
			IsFailed: false,
		},
		{
			Desc:     "int: when has the value in the middle",
			Slice:    []int{4, 2, 42, 24, 2, 4},
			Contains: 42,
			IsFailed: false,
		},

		{
			Desc:     "string: when doesn't have the element",
			Slice:    []string{"42", "24"},
			Contains: "12",
			IsFailed: true,
		},
		{
			Desc:     "string: when has the value in the beginning",
			Slice:    []string{"42", "24", "4", "2", "2", "4"},
			Contains: "42",
			IsFailed: false,
		},
		{
			Desc:     "string: when has the value includes in the end",
			Slice:    []string{"4", "2", "2", "4", "42", "24"},
			Contains: "42",
			IsFailed: false,
		},
		{
			Desc:     "string: when has the value in the middle",
			Slice:    []string{"4", "2", "42", "24", "2", "4"},
			Contains: "42",
			IsFailed: false,
		},
	} {
		t.Run(tc.Desc, AssertContainsTestCase(tc.Slice, tc.Contains, tc.IsFailed))
	}
}

func TestAsserter_Contain_sliceOfInterface(t *testing.T) {
	t.Run(`when value implements the interface`, AssertContainsTestCase([]testing.TB{t}, t, false))

	t.Run(`when value doesn't implement the interface`, func(t *testing.T) {
		assert.Must(t).Panic(func() {
			AssertContainsTestCase([]testing.TB{t}, 42, true)(t)
		})
	})
}

func TestAsserter_Contain_stringHasSub(t *testing.T) {
	type TestCase struct {
		Desc     string
		String   interface{}
		Sub      interface{}
		IsFailed bool
	}

	for _, tc := range []TestCase{
		{
			Desc:     "when doesn't have sub",
			String:   "Hello, world!",
			Sub:      "foo",
			IsFailed: true,
		},
		{
			Desc:     "when includes in the beginning",
			String:   "Hello, world!",
			Sub:      "Hello,",
			IsFailed: false,
		},
		{
			Desc:     "when includes in the middle",
			String:   "Hello, world!",
			Sub:      ", wor",
			IsFailed: false,
		},
		{
			Desc:     "when includes in the end",
			String:   "Hello, world!",
			Sub:      "world!",
			IsFailed: false,
		},
	} {
		t.Run(tc.Desc, AssertContainsTestCase(tc.String, tc.Sub, tc.IsFailed))
	}

	t.Run(`when value is a string based type`, func(t *testing.T) {
		type StringBasedType string

		t.Run(`and source has value`, AssertContainsTestCase(StringBasedType("foo/bar/baz"), StringBasedType("bar"), false))
		t.Run(`and source doesn't have value`, AssertContainsTestCase(StringBasedType("foo/bar/baz"), StringBasedType("oth"), true))
	})
}

func TestAsserter_ContainExactly_invalid(t *testing.T) {
	t.Run(`when source is invalid`, func(t *testing.T) {
		out := assert.Must(t).Panic(func() {
			asserter(func(args ...interface{}) {}).ContainExactly(nil, []int{42})
		})
		AssertFailFnArgs(t, []interface{}{"invalid expected value"}, []interface{}{out.(string)})
	})
	t.Run(`when "has" is invalid`, func(t *testing.T) {
		out := assert.Must(t).Panic(func() {
			asserter(func(args ...interface{}) {}).ContainExactly([]int{42}, nil)
		})
		AssertFailFnArgs(t, []interface{}{`invalid actual value`}, []interface{}{out.(string)})
	})
	t.Run(`invalid value asserted - nil`, func(t *testing.T) {
		assert.Must(t).Panic(func() {
			asserter(func(args ...interface{}) {}).ContainExactly([]int{42}, nil)
		})
	})
	t.Run(`non known kind is asserted`, func(t *testing.T) {
		assert.Must(t).Panic(func() {
			asserter(func(args ...interface{}) {}).ContainExactly(42, 42)
		})
	})
}

func TestAsserter_ContainExactly_map(t *testing.T) {
	type TestCase struct {
		Desc     string
		Map      interface{}
		Has      interface{}
		IsFailed bool
	}

	for _, tc := range []TestCase{
		{
			Desc:     "when equals",
			Map:      map[int]int{42: 42, 24: 24},
			Has:      map[int]int{42: 42, 24: 24},
			IsFailed: false,
		},
		{
			Desc:     "when doesn't have all the elements",
			Map:      map[int]int{42: 42, 24: 24},
			Has:      map[int]int{42: 42, 24: 24, 12: 12},
			IsFailed: true,
		},
		{
			Desc:     "when map contains sub map keys but with different value",
			Map:      map[int]int{42: 24, 24: 42},
			Has:      map[int]int{42: 42, 24: 24},
			IsFailed: true,
		},
		{
			Desc:     "when map contains sub map keys, and values are nil",
			Map:      map[int]*int{42: nil, 24: nil},
			Has:      map[int]*int{42: nil, 24: nil},
			IsFailed: false,
		},
		{
			Desc:     "when map contains sub map keys, and the key is nil",
			Map:      map[*int]int{nil: 42},
			Has:      map[*int]int{nil: 42},
			IsFailed: false,
		},
	} {
		t.Run(tc.Desc, AssertContainExactlyTestCase(tc.Map, tc.Has, tc.IsFailed))
	}
}
func TestAsserter_ContainExactly_slice(t *testing.T) {
	type TestCase struct {
		Desc     string
		Src      interface{}
		Oth      interface{}
		IsFailed bool
	}

	for _, tc := range []TestCase{
		{
			Desc:     `when elements match with order`,
			Src:      []int{42, 24},
			Oth:      []int{42, 24},
			IsFailed: false,
		},
		{
			Desc:     `when elements match without order`,
			Src:      []int{42, 24},
			Oth:      []int{24, 42},
			IsFailed: false,
		},
		{
			Desc:     `when elements do not match`,
			Src:      []int{42, 24},
			Oth:      []int{4, 2, 2, 4},
			IsFailed: true,
		},
	} {
		t.Run(tc.Desc, AssertContainExactlyTestCase(tc.Src, tc.Oth, tc.IsFailed))
	}
}

func AssertNotContainTestCase(src, has interface{}, isFailed bool) func(*testing.T) {
	return func(t *testing.T) {
		t.Helper()

		AssertContainsWith(t, isFailed, func(a assert.Asserter, msg []interface{}) {
			a.NotContain(src, has, msg...)
		})
	}
}

func TestAsserter_NotContains(t *testing.T) {
	type TestCase struct {
		Desc        string
		Source      interface{}
		NotContains interface{}
		IsFailed    bool
	}

	for _, tc := range []TestCase{
		{
			Desc:        "when slice doesn't match elements",
			Source:      []int{42, 24},
			NotContains: []int{12},
			IsFailed:    false,
		},
		{
			Desc:        "when slice contain elements",
			Source:      []int{42, 24, 12},
			NotContains: []int{24, 12},
			IsFailed:    true,
		},
		{
			Desc:        "when map doesn't contains other map elements",
			Source:      map[int]int{42: 24},
			NotContains: map[int]int{12: 6},
			IsFailed:    false,
		},
		{
			Desc:        "when map contains other map elements",
			Source:      map[int]int{42: 24, 24: 12},
			NotContains: map[int]int{24: 12},
			IsFailed:    true,
		},
		{
			Desc:        "when map contains other map's key but with different value",
			Source:      map[int]int{42: 24, 24: 12},
			NotContains: map[int]int{24: 13},
			IsFailed:    false,
		},
		{
			Desc:        "when slice doesn't include the value",
			Source:      []int{42, 24},
			NotContains: 12,
			IsFailed:    false,
		},
		{
			Desc:        "when slice does include the value",
			Source:      []int{42, 24, 12},
			NotContains: 24,
			IsFailed:    true,
		},
		{
			Desc:        "when slice of interface with map values does not have the value",
			Source:      []interface{}{map[string]int{"foo": 42}, map[string]int{}},
			NotContains: map[string]int{"bar": 42},
			IsFailed:    false,
		},
		{
			Desc:        "when slice of interface with map values has the value",
			Source:      []interface{}{map[string]int{"foo": 42}, map[string]int{}},
			NotContains: map[string]int{},
			IsFailed:    true,
		},
	} {
		t.Run(tc.Desc, AssertNotContainTestCase(tc.Source, tc.NotContains, tc.IsFailed))
	}
}

func TestAsserter_AnyOf(t *testing.T) {
	t.Run(`on happy-path`, func(t *testing.T) {
		h := assert.Must(t)
		stub := &internal.StubTB{}
		a := assert.Asserter{TB: stub, Fn: stub.Error}
		a.AnyOf(func(a *assert.AnyOf) {
			a.Test(func(it assert.It) {
				/* happy-path */
			})
			a.Test(func(it assert.It) {
				it.Must.True(false)
			})
		})
		h.Equal(false, stub.IsFailed, `testing.TB should not received any failure`)
	})

	t.Run(`on rainy-path`, func(t *testing.T) {
		h := assert.Must(t)
		stub := &internal.StubTB{}
		a := assert.Asserter{TB: stub, Fn: stub.Error}
		a.AnyOf(func(a *assert.AnyOf) {
			a.Test(func(it assert.It) {
				it.Must.True(false)
			})
		})
		h.Equal(true, stub.IsFailed, `testing.TB should failure`)
	})
}

func TestAsserter_Empty(t *testing.T) {
	type TestCase struct {
		Desc     string
		V        interface{}
		IsFailed bool
	}

	for _, tc := range []TestCase{
		{
			Desc:     "nil (for e.g.: slice before construction)",
			V:        nil,
			IsFailed: false,
		},
		{
			Desc:     "string - zero",
			V:        "",
			IsFailed: false,
		},
		{
			Desc:     "string - non zero",
			V:        "42",
			IsFailed: true,
		},
		{
			Desc:     "slice - empty",
			V:        []int{},
			IsFailed: false,
		},
		{
			Desc:     "slice - populated",
			V:        []int{42},
			IsFailed: true,
		},
		{
			Desc:     "map - empty",
			V:        map[int]int{},
			IsFailed: false,
		},
		{
			Desc:     "map - populated",
			V:        map[int]int{42: 24},
			IsFailed: true,
		},
		{
			Desc:     "array - zero values state",
			V:        [3]int{},
			IsFailed: false,
		},
		{
			Desc:     "array - populated",
			V:        [1]int{42},
			IsFailed: true,
		},
		{
			Desc:     "chan - empty",
			V:        make(chan int),
			IsFailed: false,
		},
		{
			Desc: "chan - populated",
			V: func() chan int {
				ch := make(chan int, 1)
				ch <- 42
				return ch
			}(),
			IsFailed: true,
		},
		{
			Desc:     "pointer - nil value",
			V:        (*int)(nil),
			IsFailed: false,
		},
		{
			Desc: "pointer - not nil value",
			V: func() *int {
				n := 42
				return &n
			}(),
			IsFailed: true,
		},
	} {
		tc := tc
		t.Run(tc.Desc, func(t *testing.T) {
			var failed bool
			var actualMSG []interface{}
			a := asserter(func(args ...interface{}) {
				actualMSG = args
				failed = true
			})
			expectedMSG := []interface{}{fixtures.Random.String(), fixtures.Random.Int()}
			a.Empty(tc.V, expectedMSG...)
			Equal(t, tc.IsFailed, failed)
			if failed {
				AssertFailFnArgs(t, expectedMSG, actualMSG)
			}
		})
	}
}

func TestAsserter_NotEmpty(t *testing.T) {
	type TestCase struct {
		Desc     string
		V        interface{}
		IsFailed bool
	}

	for _, tc := range []TestCase{
		{
			Desc:     "nil (for e.g.: slice before construction)",
			V:        nil,
			IsFailed: true,
		},
		{
			Desc:     "string - zero",
			V:        "",
			IsFailed: true,
		},
		{
			Desc:     "string - non zero",
			V:        "42",
			IsFailed: false,
		},
		{
			Desc:     "slice - empty",
			V:        []int{},
			IsFailed: true,
		},
		{
			Desc:     "slice - populated",
			V:        []int{42},
			IsFailed: false,
		},
		{
			Desc:     "map - empty",
			V:        map[int]int{},
			IsFailed: true,
		},
		{
			Desc:     "map - populated",
			V:        map[int]int{42: 24},
			IsFailed: false,
		},
		{
			Desc:     "array - zero values state",
			V:        [3]int{},
			IsFailed: true,
		},
		{
			Desc:     "array - populated",
			V:        [1]int{42},
			IsFailed: false,
		},
		{
			Desc:     "chan - empty",
			V:        make(chan int),
			IsFailed: true,
		},
		{
			Desc: "chan - populated",
			V: func() chan int {
				ch := make(chan int, 1)
				ch <- 42
				return ch
			}(),
			IsFailed: false,
		},
		{
			Desc:     "pointer - nil value",
			V:        (*int)(nil),
			IsFailed: true,
		},
		{
			Desc: "pointer - not nil value",
			V: func() *int {
				n := 42
				return &n
			}(),
			IsFailed: false,
		},
	} {
		tc := tc
		t.Run(tc.Desc, func(t *testing.T) {
			var failed bool
			var actualMSG []interface{}
			a := asserter(func(args ...interface{}) {
				actualMSG = args
				failed = true
			})
			expectedMSG := []interface{}{fixtures.Random.String(), fixtures.Random.Int()}
			a.NotEmpty(tc.V, expectedMSG...)
			Equal(t, tc.IsFailed, failed)
			if failed {
				AssertFailFnArgs(t, expectedMSG, actualMSG)
			}
		})
	}
}

func TestAsserter_ErrorIs(t *testing.T) {
	subject := func(tb testing.TB, expected, actual error) (failed bool) {
		var isFailed bool
		expectedMSG := []interface{}{fixtures.Random.String(), fixtures.Random.Int()}
		a := assert.Asserter{
			TB: tb,
			Fn: func(actualMSG ...interface{}) {
				AssertFailFnArgs(tb, expectedMSG, actualMSG)
				isFailed = true
			},
		}
		a.ErrorIs(expected, actual, expectedMSG...)
		return isFailed
	}

	type TestCase struct {
		Desc     string
		Expected error
		Actual   error
		IsFailed bool
	}

	exampleErr := errors.New("boom")

	for _, tc := range []TestCase{
		{
			Desc:     "when both expected and actual errors are nil, then it passes",
			Expected: nil,
			Actual:   nil,
			IsFailed: false,
		},
		{
			Desc:     "when expected is a error value, but actual is nil, then it fails",
			Expected: exampleErr,
			Actual:   nil,
			IsFailed: true,
		},
		{
			Desc:     "when expected nil, but there was actual error, then it fails",
			Expected: nil,
			Actual:   exampleErr,
			IsFailed: true,
		},
		{
			Desc:     "when expected an error is the same as the actual error, then it passes",
			Expected: exampleErr,
			// intentionally different errors.errorString with the same value
			Actual:   errors.New(exampleErr.Error()),
			IsFailed: false,
		},
		{
			Desc:     "when expected an error, and the actual error wraps is, then it passes",
			Expected: exampleErr,
			Actual:   fmt.Errorf("wrapped error: %w", exampleErr),
			IsFailed: false,
		},
	} {
		tc := tc
		t.Run(tc.Desc, func(t *testing.T) {
			Equal(t, tc.IsFailed, subject(t, tc.Expected, tc.Actual))
		})
	}
}

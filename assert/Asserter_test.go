package assert_test

import (
	"context"
	"errors"
	"fmt"
	"io"
	"math/big"
	"net"
	"reflect"
	"strings"
	"sync/atomic"
	"testing"
	"testing/iotest"
	"time"

	"github.com/adamluzsi/testcase/internal/doubles"
	"github.com/adamluzsi/testcase/sandbox"

	"github.com/adamluzsi/testcase/assert"
	"github.com/adamluzsi/testcase/random"
)

func TestAsserter_Equal_pp(t *testing.T) {
	dtb := &doubles.TB{}
	type X struct{ A, B, C int }
	assert.Should(dtb).Equal(X{A: 1, B: 2, C: 3}, X{A: 1, B: 42, C: 3})
	t.Log(dtb.Logs.String())
}

func TestMust(t *testing.T) {
	must := assert.Must(t)
	stub := &doubles.TB{}
	out := sandbox.Run(func() {
		a := assert.Must(stub)
		a.True(false) // fail it
		t.Fail()
	})
	must.False(out.OK, "failed while stopping the goroutine")
	must.True(stub.IsFailed)
}

func TestShould(t *testing.T) {
	must := assert.Must(t)
	stub := &doubles.TB{}
	out := sandbox.Run(func() {
		a := assert.Should(stub)
		a.True(false) // fail it
	})
	must.True(out.OK)
	must.False(out.Goexit, "failed without stopping the goroutine")
	must.True(stub.IsFailed)
}

func asserter(dtb *doubles.TB) assert.Asserter {
	return assert.Should(dtb)
}

func Equal(tb testing.TB, a, b interface{}) {
	tb.Helper()
	if !reflect.DeepEqual(a, b) {
		tb.Fatalf("A and B not equal: %#v <=> %#v", a, b)
	}
}

func Contain(tb testing.TB, haystack, needle string) {
	tb.Helper()
	if !strings.Contains(haystack, needle) {
		tb.Fatalf("\nhaystack: %#v\nneedle: %#v\n", haystack, needle)
	}
}

func TestAsserter_True(t *testing.T) {
	t.Run(`when true passed`, func(t *testing.T) {
		dtb := &doubles.TB{}
		subject := asserter(dtb)
		subject.True(true)
		Equal(t, dtb.IsFailed, false)
	})
	t.Run(`when false passed`, func(t *testing.T) {
		dtb := &doubles.TB{}
		subject := asserter(dtb)
		expectedMsg := []interface{}{"foo", "bar", "baz"}
		subject.True(false, expectedMsg...)
		Equal(t, dtb.IsFailed, true)
		AssertFailMsg(t, dtb, expectedMsg)
	})
}

func AssertFailMsg(tb testing.TB, dtb *doubles.TB, msgs []any) {
	tb.Helper()
	logs := dtb.Logs.String()
	for _, msg := range msgs {
		Contain(tb, logs, strings.TrimSpace(fmt.Sprint(msg)))
	}
}

func TestAsserter_False(t *testing.T) {
	t.Run(`when true passed`, func(t *testing.T) {
		dtb := &doubles.TB{}
		subject := asserter(dtb)
		expectedMsg := []interface{}{"hello", "world", 42}
		subject.False(true, expectedMsg...)
		Equal(t, dtb.IsFailed, true)
		AssertFailMsg(t, dtb, expectedMsg)
	})
	t.Run(`when false passed`, func(t *testing.T) {
		dtb := &doubles.TB{}
		subject := asserter(dtb)
		subject.False(false)
		Equal(t, dtb.IsFailed, false)
	})
}

func TestAsserter_Nil(t *testing.T) {
	t.Run(`when nil passed, then it is accepted`, func(t *testing.T) {
		dtb := &doubles.TB{}
		subject := asserter(dtb)
		subject.Nil(nil)
		Equal(t, dtb.IsFailed, false)
	})
	t.Run(`when pointer with nil value passed, then it is accepted as nil`, func(t *testing.T) {
		dtb := &doubles.TB{}
		subject := asserter(dtb)
		subject.Nil((func())(nil))
		Equal(t, dtb.IsFailed, false)
	})
	t.Run(`when non nil value is passed`, func(t *testing.T) {
		dtb := &doubles.TB{}
		subject := asserter(dtb)
		expectedMsg := []interface{}{"foo", "bar", "baz"}
		subject.Nil(errors.New("not nil"), expectedMsg...)
		Equal(t, dtb.IsFailed, true)
		AssertFailMsg(t, dtb, expectedMsg)
	})
	t.Run("when non nil zero value is passed", func(t *testing.T) {
		dtb := &doubles.TB{}
		subject := asserter(dtb)
		expectedMsg := []interface{}{"foo", "bar", "baz"}
		subject.Nil("", expectedMsg...) // zero string value
		Equal(t, dtb.IsFailed, true)
		AssertFailMsg(t, dtb, expectedMsg)
	})
}

func TestAsserter_NotNil(t *testing.T) {
	t.Run(`when nil passed`, func(t *testing.T) {
		dtb := &doubles.TB{}
		subject := asserter(dtb)
		msg := []interface{}{"foo", "bar", "baz"}
		subject.NotNil(nil, msg...)
		AssertFailMsg(t, dtb, msg)
	})
	t.Run(`when pointer with nil value passed, then it is refused as nil`, func(t *testing.T) {
		dtb := &doubles.TB{}
		subject := asserter(dtb)
		subject.NotNil((func())(nil))
		Equal(t, dtb.IsFailed, true)
	})
	t.Run(`when non nil value is passed`, func(t *testing.T) {
		dtb := &doubles.TB{}
		subject := asserter(dtb)
		subject.NotNil(errors.New("not nil"), "foo", "bar", "baz")
		Equal(t, dtb.IsFailed, false)
	})
	t.Run("when non nil zero value is passed", func(t *testing.T) {
		dtb := &doubles.TB{}
		subject := asserter(dtb)
		subject.NotNil("", "foo", "bar", "baz")
		Equal(t, dtb.IsFailed, false)
	})
}

func TestAsserter_Panics(t *testing.T) {
	t.Run(`when no panic, fails`, func(t *testing.T) {
		dtb := &doubles.TB{}
		subject := asserter(dtb)
		subject.Panic(func() { /* nothing */ }, "boom!")
		Equal(t, dtb.IsFailed, true)
		AssertFailMsg(t, dtb, []interface{}{"boom!"})
	})
	t.Run(`when panic with nil value, pass`, func(t *testing.T) {
		dtb := &doubles.TB{}
		subject := asserter(dtb)
		subject.Panic(func() { panic(nil) }, "boom!")
		Equal(t, dtb.IsFailed, false)
	})
	t.Run(`when panic with something, pass`, func(t *testing.T) {
		dtb := &doubles.TB{}
		subject := asserter(dtb)
		subject.Panic(func() { panic("something") }, "boom!")
		Equal(t, dtb.IsFailed, false)
	})
}

func TestAsserter_NotPanics(t *testing.T) {
	t.Run(`when no panic, pass`, func(t *testing.T) {
		dtb := &doubles.TB{}
		subject := asserter(dtb)
		subject.NotPanic(func() { /* nothing */ }, "boom!")
		Equal(t, dtb.IsFailed, false)
	})
	t.Run(`when panic with nil value, fail`, func(t *testing.T) {
		dtb := &doubles.TB{}
		subject := asserter(dtb)
		subject.NotPanic(func() { panic(nil) }, "boom!")
		Equal(t, dtb.IsFailed, true)
		AssertFailMsg(t, dtb, []interface{}{"boom!"})
	})
	t.Run(`when panic with something, fail`, func(t *testing.T) {
		dtb := &doubles.TB{}
		subject := asserter(dtb)
		subject.NotPanic(func() { panic("something") }, "boom!")
		Equal(t, dtb.IsFailed, true)
		AssertFailMsg(t, dtb, []interface{}{"boom!"})
		AssertFailMsg(t, dtb, []interface{}{"something"})
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

	var expFunc = func() {}

	rnd := random.New(random.CryptoSeed{})
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
		{
			Desc:     "when func provided - equal - no error",
			Expected: expFunc,
			Actual:   expFunc,
			IsFailed: false,
		},
		{
			Desc: "when value implements equalable and the two value is equal by IsEqual",
			Expected: ExampleEqualableWithIsEqual{
				relevantUnexportedValue: 42,
				IrrelevantExportedField: 42,
			},
			Actual: ExampleEqualableWithIsEqual{
				relevantUnexportedValue: 42,
				IrrelevantExportedField: 24,
			},
			IsFailed: false,
		},
		{
			Desc: "when value implements equalable and the two value is not equal by IsEqual",
			Expected: ExampleEqualableWithIsEqual{
				relevantUnexportedValue: 24,
				IrrelevantExportedField: 42,
			},
			Actual: ExampleEqualableWithIsEqual{
				relevantUnexportedValue: 42,
				IrrelevantExportedField: 42,
			},
			IsFailed: true,
		},
		{
			Desc: "when value implements equalableWithError and the two value is equal by IsEqual",
			Expected: ExampleEqualableWithError{
				relevantUnexportedValue: 42,
				IrrelevantExportedField: 42,
			},
			Actual: ExampleEqualableWithError{
				relevantUnexportedValue: 42,
				IrrelevantExportedField: 4242,
			},
			IsFailed: false,
		},
		{
			Desc: "when value implements equalableWithError and the two value is not equal by IsEqual",
			Expected: ExampleEqualableWithError{
				relevantUnexportedValue: 42,
				IrrelevantExportedField: 42,
			},
			Actual: ExampleEqualableWithError{
				relevantUnexportedValue: 4242,
				IrrelevantExportedField: 42,
			},
			IsFailed: true,
		},
		{
			Desc:     "when nil slice compared to an empty slice",
			Expected: ([]int)(nil),
			Actual:   []int{},
			IsFailed: true,
		},
	} {
		tc := tc
		t.Run(tc.Desc, func(t *testing.T) {
			t.Logf("expected: %#v", tc.Expected)
			t.Logf("actual: %#v", tc.Actual)

			expectedMsg := []interface{}{rnd.StringN(3), rnd.StringN(3)}
			dtb := &doubles.TB{}
			subject := asserter(dtb)
			subject.Equal(tc.Expected, tc.Actual, expectedMsg...)
			Equal(t, dtb.IsFailed, tc.IsFailed)
			if tc.IsFailed {
				AssertFailMsg(t, dtb, expectedMsg)
			}
		})
	}
}

func TestAsserter_Equal_typeSafety(t *testing.T) {
	type (
		MainType string
		SubType  MainType
	)
	t.Run("when types and values are the same", func(t *testing.T) {
		dtb := &doubles.TB{}
		assert.Should(dtb).Equal(SubType("A"), SubType("A"))
		assert.False(t, dtb.Failed())
	})
	t.Run("when types the same but values are different", func(t *testing.T) {
		dtb := &doubles.TB{}
		assert.Should(dtb).Equal(SubType("A"), SubType("B"))
		assert.True(t, dtb.Failed())
	})
	t.Run("when types are different but values are the same", func(t *testing.T) {
		dtb := &doubles.TB{}
		assert.Should(dtb).Equal(MainType("A"), SubType("A"))
		assert.True(t, dtb.Failed())
		assert.Contain(t, dtb.Logs.String(), "incorrect type")
	})
}

func TestAsserter_Equal_types(t *testing.T) {
	rnd := random.New(random.CryptoSeed{})
	t.Run("time.Time", func(t *testing.T) {
		t.Run("OK", func(t *testing.T) {
			v := time.Now()
			dtb := &doubles.TB{}
			out := sandbox.Run(func() { assert.Equal(dtb, v, v.UTC()) })
			Equal(t, false, dtb.IsFailed)
			Equal(t, true, out.OK)
		})
		t.Run("NOK", func(t *testing.T) {
			v := time.Now()
			dtb := &doubles.TB{}
			out := sandbox.Run(func() { assert.Equal(dtb, v, v.Add(time.Second)) })
			Equal(t, true, dtb.IsFailed)
			Equal(t, false, out.OK)
		})
		t.Run("as embedded value", func(t *testing.T) {
			type T struct{ V time.Time }
			var (
				dtb = &doubles.TB{}
				tm1 = time.Now().UTC()
				tm2 = tm1.Local()
				v1  = T{V: tm1}
				v2  = T{V: tm2}
			)
			out := sandbox.Run(func() {
				assert.Equal(dtb, v1, v2)
			})
			Equal(t, false, dtb.IsFailed)
			Equal(t, true, out.OK)
		})
	})
	t.Run("net.IP", func(t *testing.T) {
		t.Run("OK", func(t *testing.T) {
			ip1 := net.ParseIP("192.0.2.1")
			ip2 := net.ParseIP("::ffff:192.0.2.1")
			dtb := &doubles.TB{}
			out := sandbox.Run(func() { assert.Equal(dtb, ip1, ip2) })
			Equal(t, false, dtb.IsFailed)
			Equal(t, true, out.OK)
		})
		t.Run("NOK", func(t *testing.T) {
			ip1 := net.ParseIP("192.0.2.2")
			ip2 := net.ParseIP("::ffff:192.0.2.1")
			dtb := &doubles.TB{}
			out := sandbox.Run(func() { assert.Equal(dtb, ip1, ip2) })
			Equal(t, true, dtb.IsFailed)
			Equal(t, false, out.OK)
		})
	})
	t.Run("big.Int", func(t *testing.T) {
		t.Run("OK", func(t *testing.T) {
			v := int64(rnd.IntB(128, 1024))
			// Create two *big.Int values
			bi1 := big.NewInt(v)
			bi2 := big.NewInt(v * -1)
			bi2.Neg(bi2) // Negate bi2
			dtb := &doubles.TB{}
			out := sandbox.Run(func() { assert.Equal(dtb, bi1, bi2) })
			Equal(t, false, dtb.IsFailed)
			Equal(t, true, out.OK)
		})
		t.Run("NOK", func(t *testing.T) {
			bi1 := big.NewInt(int64(rnd.IntB(128, 1024)))
			bi2 := big.NewInt(int64(rnd.IntB(128, 1024)))
			dtb := &doubles.TB{}
			out := sandbox.Run(func() { assert.Equal(dtb, bi1, bi2) })
			Equal(t, true, dtb.IsFailed)
			Equal(t, false, out.OK)
		})
	})
	t.Run("big.Float", func(t *testing.T) {
		t.Run("OK", func(t *testing.T) {
			v := rnd.Float64()
			bi1 := big.NewFloat(v)
			bi2 := big.NewFloat(v)
			dtb := &doubles.TB{}
			out := sandbox.Run(func() { assert.Equal(dtb, bi1, bi2) })
			Equal(t, false, dtb.IsFailed)
			Equal(t, true, out.OK)
		})
		t.Run("NOK", func(t *testing.T) {
			bi1 := big.NewFloat(rnd.Float64())
			bi2 := big.NewFloat(rnd.Float64())
			dtb := &doubles.TB{}
			out := sandbox.Run(func() { assert.Equal(dtb, bi1, bi2) })
			Equal(t, true, dtb.IsFailed)
			Equal(t, false, out.OK)
		})
	})
	t.Run("big.Rat", func(t *testing.T) {
		t.Run("OK", func(t *testing.T) {
			a, b := int64(rnd.IntB(128, 256)), int64(rnd.IntB(0, 42))
			bi1 := big.NewRat(a, b)
			bi2 := big.NewRat(a, b)
			dtb := &doubles.TB{}
			out := sandbox.Run(func() { assert.Equal(dtb, bi1, bi2) })
			Equal(t, false, dtb.IsFailed)
			Equal(t, true, out.OK)
		})
		t.Run("NOK", func(t *testing.T) {
			bi1 := big.NewRat(int64(rnd.IntB(128, 256)), int64(rnd.IntB(0, 42)))
			bi2 := big.NewRat(int64(rnd.IntB(128, 256)), int64(rnd.IntB(0, 42)))
			dtb := &doubles.TB{}
			out := sandbox.Run(func() { assert.Equal(dtb, bi1, bi2) })
			Equal(t, true, dtb.IsFailed)
			Equal(t, false, out.OK)
		})
	})
}

func TestAsserter_Equal_equalableWithError_ErrorReturned(t *testing.T) {
	t.Log("when value implements equalableWithError and IsEqual returns an error")

	expected := ExampleEqualableWithError{
		relevantUnexportedValue: 42,
		IrrelevantExportedField: 42,
		IsEqualErr:              errors.New("boom"),
	}

	actual := ExampleEqualableWithError{
		relevantUnexportedValue: 42,
		IrrelevantExportedField: 42,
	}

	stub := &doubles.TB{}

	sandbox.Run(func() {
		a := assert.Asserter{
			TB:   stub,
			Fail: stub.FailNow,
		}

		a.Equal(expected, actual)
	})
	if !stub.IsFailed {
		t.Fatal("expected that testing.TB is failed because the returned error")
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

	rnd := random.New(random.CryptoSeed{})

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
			expectedMsg := []interface{}{rnd.StringN(3), rnd.StringN(3)}
			dtb := &doubles.TB{}
			subject := asserter(dtb)
			subject.NotEqual(tc.Expected, tc.Actual, expectedMsg...)
			Equal(t, dtb.IsFailed, tc.IsFailed)
			if tc.IsFailed {
				AssertFailMsg(t, dtb, expectedMsg)
			}
		})
	}
}

func TestAsserter_NotEqual_typeSafety(t *testing.T) {
	type (
		MainType string
		SubType  MainType
	)
	t.Run("when types are the same and values are different", func(t *testing.T) {
		dtb := &doubles.TB{}
		assert.Should(dtb).NotEqual(SubType("A"), SubType("B"))
		assert.False(t, dtb.Failed())
	})
	t.Run("when types are the same and values are the same", func(t *testing.T) {
		dtb := &doubles.TB{}
		assert.Should(dtb).NotEqual(SubType("A"), SubType("A"))
		assert.True(t, dtb.Failed())
	})
	t.Run("when types and values are different", func(t *testing.T) {
		dtb := &doubles.TB{}
		assert.Should(dtb).NotEqual(MainType("A"), SubType("B"))
		assert.True(t, dtb.Failed())
		assert.Contain(t, dtb.Logs.String(), "incorrect type")
	})
}

func AssertContainsWith(tb testing.TB, isFailed bool, contains func(a assert.Asserter, msg []interface{})) {
	tb.Helper()
	rnd := random.New(random.CryptoSeed{})
	expectedMsg := []interface{}{rnd.StringN(3), rnd.StringN(3)}
	dtb := &doubles.TB{}
	subject := asserter(dtb)
	contains(subject, expectedMsg)
	Equal(tb, dtb.IsFailed, isFailed)
	if isFailed {
		AssertFailMsg(tb, dtb, expectedMsg)
	}
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
		dtb := &doubles.TB{}
		asserter(dtb).Contain(nil, []int{42})
		AssertFailMsg(t, dtb, []interface{}{"invalid source value"})
	})
	t.Run(`when "has" is invalid`, func(t *testing.T) {
		dtb := &doubles.TB{}
		asserter(dtb).Contain([]int{42}, nil)
		AssertFailMsg(t, dtb, []interface{}{`invalid "has" value`})
	})
}

func TestAsserter_Contain_typeMismatch(t *testing.T) {
	assert.Must(t).Panic(func() {
		dtb := &doubles.TB{}
		asserter(dtb).Contain([]int{42}, []string{"42"})
	}, "will panic on type mismatch")
	assert.Must(t).Panic(func() {
		dtb := &doubles.TB{}
		asserter(dtb).Contain([]int{42}, "42")
	}, "will panic on type mismatch")
}

func TestAsserter_Contain_sliceHasSubSlice(t *testing.T) {
	type TestCase struct {
		Desc     string
		Haystack interface{}
		Needle   interface{}
		IsFailed bool
	}

	for _, tc := range []TestCase{
		{
			Desc:     "[]int: when equals",
			Haystack: []int{42, 24},
			Needle:   []int{42, 24},
			IsFailed: false,
		},
		{
			Desc:     "[]int: when needle doesn't have all the elements of the haystack",
			Haystack: []int{42, 24},
			Needle:   []int{42, 24, 42},
			IsFailed: true,
		},
		{
			Desc:     "[]int: when fully includes it in the beginning",
			Haystack: []int{42, 24, 4, 2, 2, 4},
			Needle:   []int{42, 24},
			IsFailed: false,
		},
		{
			Desc:     "[]int: when fully includes it in the end",
			Haystack: []int{4, 2, 2, 4, 42, 24},
			Needle:   []int{42, 24},
			IsFailed: false,
		},
		{
			Desc:     "[]int: when fully includes it in the middle",
			Haystack: []int{4, 2, 42, 24, 2, 4},
			Needle:   []int{42, 24},
			IsFailed: false,
		},
		{
			Desc:     "[]int: when fully includes it from random sections",
			Haystack: []int{4, 2, 42, 24, 2, 4},
			Needle:   []int{2, 24, 42},
			IsFailed: false,
		},
		{
			Desc:     "[]string: when equals",
			Haystack: []string{"42", "24"},
			Needle:   []string{"42", "24"},
			IsFailed: false,
		},
		{
			Desc:     "[]string: when haystack slice has less element that the needle slice",
			Haystack: []string{"42", "24"},
			Needle:   []string{"42", "24", "42"},
			IsFailed: true,
		},
		{
			Desc:     "[]string: when doesn't have fully matching elements",
			Haystack: []string{"42", "42"},
			Needle:   []string{"42", "41"},
			IsFailed: true,
		},
		{
			Desc:     "[]string: when fully includes it in the beginning",
			Haystack: []string{"42", "24", "4", "2", "2", "4"},
			Needle:   []string{"42", "24"},
			IsFailed: false,
		},
		{
			Desc:     "[]string: when fully includes it in the end",
			Haystack: []string{"4", "2", "2", "4", "42", "24"},
			Needle:   []string{"42", "24"},
			IsFailed: false,
		},
		{
			Desc:     "[]string: when fully includes it in the middle",
			Haystack: []string{"4", "2", "42", "24", "2", "4"},
			Needle:   []string{"42", "24"},
			IsFailed: false,
		},
		{
			Desc:     "[]string: when fully includes it in the middle",
			Haystack: []string{"4", "2", "42", "24", "2", "4"},
			Needle:   []string{"42", "24"},
			IsFailed: false,
		},
	} {
		t.Run(tc.Desc, AssertContainsTestCase(tc.Haystack, tc.Needle, tc.IsFailed))
	}
}

func TestAsserter_Subset(t *testing.T) {
	type TestCase struct {
		Desc     string
		Slice    interface{}
		Subset   interface{}
		IsFailed bool
	}

	for _, tc := range []TestCase{
		{
			Desc:     "[]int: when equals",
			Slice:    []int{42, 24},
			Subset:   []int{42, 24},
			IsFailed: false,
		},
		{
			Desc:     "[]int: when doesn't have all the elements",
			Slice:    []int{42, 24},
			Subset:   []int{42, 24, 42},
			IsFailed: true,
		},
		{
			Desc:     "[]int: when fully includes in the beginning",
			Slice:    []int{42, 24, 4, 2, 2, 4},
			Subset:   []int{42, 24},
			IsFailed: false,
		},
		{
			Desc:     "[]int: when fully includes in the end",
			Slice:    []int{4, 2, 2, 4, 42, 24},
			Subset:   []int{42, 24},
			IsFailed: false,
		},
		{
			Desc:     "[]int: when fully includes in the middle",
			Slice:    []int{4, 2, 42, 24, 2, 4},
			Subset:   []int{42, 24},
			IsFailed: false,
		},
		{
			Desc:     "[]string: when equals",
			Slice:    []string{"42", "24"},
			Subset:   []string{"42", "24"},
			IsFailed: false,
		},
		{
			Desc:     "[]string: when slice has less element that the sub slice",
			Slice:    []string{"42", "24"},
			Subset:   []string{"42", "24", "42"},
			IsFailed: true,
		},
		{
			Desc:     "[]string: when doesn't have fully matching elements",
			Slice:    []string{"42", "42"},
			Subset:   []string{"42", "41"},
			IsFailed: true,
		},
		{
			Desc:     "[]string: when fully includes in the beginning",
			Slice:    []string{"42", "24", "4", "2", "2", "4"},
			Subset:   []string{"42", "24"},
			IsFailed: false,
		},
		{
			Desc:     "[]string: when fully includes in the end",
			Slice:    []string{"4", "2", "2", "4", "42", "24"},
			Subset:   []string{"42", "24"},
			IsFailed: false,
		},
		{
			Desc:     "[]string: when fully includes in the middle",
			Slice:    []string{"4", "2", "42", "24", "2", "4"},
			Subset:   []string{"42", "24"},
			IsFailed: false,
		},
		{
			Desc:     "[]string: when fully includes in the middle",
			Slice:    []string{"4", "2", "42", "24", "2", "4"},
			Subset:   []string{"42", "24"},
			IsFailed: false,
		},
	} {
		t.Run(tc.Desc, func(t *testing.T) {
			AssertContainsWith(t, tc.IsFailed, func(a assert.Asserter, msg []interface{}) {
				a.Sub(tc.Slice, tc.Subset, msg...)
			})
		})
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
			asserter(&doubles.TB{}).ContainExactly(nil, []int{42})
		})
		Contain(t, out.(string), "invalid expected value")
	})
	t.Run(`when "has" is invalid`, func(t *testing.T) {
		out := assert.Must(t).Panic(func() {
			asserter(&doubles.TB{}).ContainExactly([]int{42}, nil)
		})
		Contain(t, out.(string), `invalid actual value`)
	})
	t.Run(`invalid value asserted - nil`, func(t *testing.T) {
		assert.Must(t).Panic(func() {
			asserter(&doubles.TB{}).ContainExactly([]int{42}, nil)
		})
	})
	t.Run(`non known kind is asserted`, func(t *testing.T) {
		assert.Must(t).Panic(func() {
			asserter(&doubles.TB{}).ContainExactly(42, 42)
		})
	})
}

func TestAsserter_Within(t *testing.T) {
	t.Run("when block finish within time", func(t *testing.T) {
		dtb := &doubles.TB{}
		a := asserter(dtb)
		a.Within(time.Second, func(ctx context.Context) {
			select {
			case <-time.After(time.Microsecond):
			case <-ctx.Done():
			}
		})
		assert.False(t, dtb.IsFailed)
	})
	t.Run("when block takes more time than the accepted timeout, assertion fails", func(t *testing.T) {
		dtb := &doubles.TB{}
		a := asserter(dtb)
		a.Within(time.Microsecond, func(ctx context.Context) {
			timer := time.NewTimer(time.Hour)
			select {
			case <-timer.C:
			case <-ctx.Done():
				timer.Stop()
			}
		})
		assert.True(t, dtb.IsFailed)
	})
	t.Run("when block takes more time than the accepted timeout, the function's context is cancelled", func(t *testing.T) {
		dtb := &doubles.TB{}
		a := asserter(dtb)

		var isCancelled int32
		a.Within(time.Microsecond, func(ctx context.Context) {
			timer := time.NewTimer(time.Second)
			select {
			case <-timer.C:
			case <-ctx.Done():
				timer.Stop()
				atomic.AddInt32(&isCancelled, 1)
			}
		})
		assert.True(t, dtb.IsFailed)

		assert.EventuallyWithin(3*time.Second).Assert(t, func(it assert.It) {
			it.Must.True(atomic.LoadInt32(&isCancelled) == 1)
		})
	})
	t.Run("when FailNow based failing as part of the Within block, it is propagated to the outside as well", func(t *testing.T) {
		dtb := &doubles.TB{}
		a := asserter(dtb)

		ro := sandbox.Run(func() {
			a.Within(time.Second, func(ctx context.Context) {
				dtb.FailNow()
			})
		})

		assert.Must(t).True(dtb.IsFailed)
		assert.Must(t).True(ro.Goexit)
	})
}

func TestAsserter_NotWithin(t *testing.T) {
	t.Run("when block finish within time, then assert fails", func(t *testing.T) {
		dtb := &doubles.TB{}
		a := asserter(dtb)
		a.NotWithin(time.Second, func(ctx context.Context) {
			select {
			case <-time.After(time.Microsecond):
			case <-ctx.Done():
			}
		})
		assert.True(t, dtb.IsFailed)
	})
	t.Run("when block takes more time than the timeout, assertion succeed", func(t *testing.T) {
		dtb := &doubles.TB{}
		a := asserter(dtb)
		a.NotWithin(time.Microsecond, func(ctx context.Context) {
			timer := time.NewTimer(time.Hour)
			select {
			case <-timer.C:
			case <-ctx.Done():
				timer.Stop()
			}
		})
		assert.False(t, dtb.IsFailed)
	})
	t.Run("when block takes more time than the timeout, the function's context is cancelled", func(t *testing.T) {
		dtb := &doubles.TB{}
		a := asserter(dtb)

		var isCancelled int32
		a.NotWithin(time.Microsecond, func(ctx context.Context) {
			timer := time.NewTimer(time.Second)
			select {
			case <-timer.C:
			case <-ctx.Done():
				timer.Stop()
				atomic.AddInt32(&isCancelled, 1)
			}
		})
		assert.False(t, dtb.IsFailed)

		assert.EventuallyWithin(3*time.Second).Assert(t, func(it assert.It) {
			it.Must.True(atomic.LoadInt32(&isCancelled) == 1)
		})
	})
	t.Run("when FailNow based failing as part of the Within block, it is propagated to the outside as well", func(t *testing.T) {
		dtb := &doubles.TB{}
		a := asserter(dtb)

		ro := sandbox.Run(func() {
			a.NotWithin(time.Second, func(ctx context.Context) {
				dtb.FailNow()
			})
		})

		assert.Must(t).True(dtb.IsFailed)
		assert.Must(t).True(ro.Goexit)
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
		{
			Desc:     "when slices has matching values, but the other slice has additional value as well",
			Src:      []string{"42", "24"},
			Oth:      []string{"24", "42", "13"},
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
		stub := &doubles.TB{}
		a := assert.Asserter{TB: stub, Fail: stub.Fail}
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
		stub := &doubles.TB{}
		a := assert.Asserter{TB: stub, Fail: stub.Fail}
		a.AnyOf(func(a *assert.AnyOf) {
			a.Test(func(it assert.It) {
				it.Must.True(false)
			})
		})
		h.Equal(true, stub.IsFailed, `testing.TB should failure`)
	})

	t.Run("on rainy path - when test fails during AnyOf.Finish, then outer layer also fails", func(t *testing.T) {
		ro := sandbox.Run(func() {
			stub := &doubles.TB{}
			a := assert.Asserter{TB: stub, Fail: stub.FailNow}
			a.AnyOf(func(a *assert.AnyOf) {
				a.Test(func(it assert.It) { it.FailNow() })
			})
		})
		t.Log(`Asserter was used with FailNow, so the sandbox should not be OK`)
		assert.False(t, ro.OK)
		assert.True(t, ro.Goexit)
		assert.Nil(t, ro.PanicValue)
	})
}

func TestAsserter_Empty(t *testing.T) {
	type TestCase struct {
		Desc     string
		V        interface{}
		IsFailed bool
	}

	rnd := random.New(random.CryptoSeed{})
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
		{
			Desc:     "time - not zero value",
			V:        time.Now(),
			IsFailed: true,
		},
		{
			Desc:     "time - zero value",
			V:        time.Time{},
			IsFailed: false,
		},
		{
			Desc:     "time - zero value with different time zone",
			V:        time.Time{}.UTC().Local(),
			IsFailed: false,
		},
	} {
		tc := tc
		t.Run(tc.Desc, func(t *testing.T) {
			dtb := &doubles.TB{}
			a := asserter(dtb)
			expectedMSG := []interface{}{rnd.String(), rnd.Int()}
			a.Empty(tc.V, expectedMSG...)
			Equal(t, tc.IsFailed, dtb.IsFailed)
			if tc.IsFailed {
				AssertFailMsg(t, dtb, expectedMSG)
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

	rnd := random.New(random.CryptoSeed{})
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
		{
			Desc:     "time - not zero value",
			V:        time.Now(),
			IsFailed: false,
		},
		{
			Desc:     "time - zero value",
			V:        time.Time{},
			IsFailed: true,
		},
		{
			Desc:     "time - zero value with different time zone",
			V:        time.Time{}.UTC().Local(),
			IsFailed: true,
		},
	} {
		tc := tc
		t.Run(tc.Desc, func(t *testing.T) {
			dtb := &doubles.TB{}
			a := asserter(dtb)
			expectedMSG := []interface{}{rnd.String(), rnd.Int()}
			a.NotEmpty(tc.V, expectedMSG...)
			Equal(t, tc.IsFailed, dtb.IsFailed)
			if tc.IsFailed {
				AssertFailMsg(t, dtb, expectedMSG)
			}
		})
	}
}

func TestAsserter_ErrorIs(t *testing.T) {
	rnd := random.New(random.CryptoSeed{})
	subject := func(tb testing.TB, isFailed bool, expected, actual error) (failed bool) {
		expectedMSG := []interface{}{rnd.String(), rnd.Int()}
		dtb := &doubles.TB{}
		a := assert.Asserter{TB: dtb, Fail: dtb.Fail}
		a.ErrorIs(expected, actual, expectedMSG...)
		if isFailed {
			AssertFailMsg(t, dtb, expectedMSG)
		}
		return dtb.Failed()
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
			Desc:     "when expected an error is the same as the actual error, and also wrapped, then it passes",
			Expected: exampleErr,
			// intentionally different errors.errorString with the same value
			Actual:   fmt.Errorf("%w", errors.New(exampleErr.Error())),
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
			Equal(t, tc.IsFailed, subject(t, tc.IsFailed, tc.Expected, tc.Actual))
		})
	}
}

func TestAsserter_Error(t *testing.T) {
	t.Run(`when nil passed, then it is fails`, func(t *testing.T) {
		dtb := &doubles.TB{}
		subject := asserter(dtb)

		expectedMsg := []interface{}{"foo", "bar", "baz"}
		subject.Error(nil, expectedMsg...)
		Equal(t, dtb.IsFailed, true)
		AssertFailMsg(t, dtb, expectedMsg)
	})

	t.Run(`when non-nil error value is passed, then it is accepted`, func(t *testing.T) {
		dtb := &doubles.TB{}
		subject := asserter(dtb)
		subject.Error(errors.New("boom"))
		Equal(t, dtb.IsFailed, false)
	})
}

func TestAsserter_NoError(t *testing.T) {
	t.Run(`when nil passed, then it is accepted`, func(t *testing.T) {
		dtb := &doubles.TB{}
		subject := asserter(dtb)
		subject.NoError(nil)
		Equal(t, dtb.IsFailed, false)
	})
	t.Run(`when non-nil error value is passed`, func(t *testing.T) {
		dtb := &doubles.TB{}
		subject := asserter(dtb)
		expectedMsg := []interface{}{"foo", "bar", "baz"}
		subject.NoError(errors.New("boom"), expectedMsg...)
		Equal(t, dtb.IsFailed, true)
		AssertFailMsg(t, dtb, expectedMsg)
	})
}

func TestAsserter_Read(t *testing.T) {
	type TestCase struct {
		Desc     string
		Expected any
		Reader   io.Reader
		Failed   bool
	}
	const sampleText = "Hello, world!"
	for _, tc := range []TestCase{
		{
			Desc:     "when value of the reader is nil",
			Expected: sampleText,
			Reader:   nil,
			Failed:   true,
		},
		{
			Desc:     "when string expectation is matching with reader's content",
			Expected: string(sampleText),
			Reader:   strings.NewReader(sampleText),
			Failed:   false,
		},
		{
			Desc:     "when string expectation is different from reader's content",
			Expected: string(sampleText + ", but different"),
			Reader:   strings.NewReader(sampleText),
			Failed:   true,
		},
		{
			Desc:     "when []byte expectation is matching with reader's content",
			Expected: []byte(sampleText),
			Reader:   strings.NewReader(sampleText),
			Failed:   false,
		},
		{
			Desc:     "when []byte expectation is different from reader's content",
			Expected: []byte(sampleText + ", but different"),
			Reader:   strings.NewReader(sampleText),
			Failed:   true,
		},
	} {
		tc := tc
		t.Run(tc.Desc, func(t *testing.T) {
			dtb := &doubles.TB{}
			subject := asserter(dtb)
			msg := []any{"asd", "dsa"}
			subject.Read(tc.Expected, tc.Reader, msg...)
			Equal(t, tc.Failed, dtb.IsFailed)
			if tc.Failed {
				AssertFailMsg(t, dtb, msg)
			}
		})
	}
}

func TestAsserter_ReadAll(t *testing.T) {
	rnd := random.New(random.CryptoSeed{})
	msg := []any{rnd.String(), rnd.String()}
	logMSG := strings.TrimSpace(fmt.Sprintln(msg...))
	t.Run("when io.Reader has readable content", func(t *testing.T) {
		dtb := &doubles.TB{}
		subject := asserter(dtb)
		txt := rnd.String()
		data := subject.ReadAll(strings.NewReader(txt))
		Equal(t, dtb.IsFailed, false)
		Equal(t, txt, string(data))
	})
	t.Run("when io.Reader is nil", func(t *testing.T) {
		dtb := &doubles.TB{}
		subject := asserter(dtb)
		_ = subject.ReadAll(nil, msg...)
		Equal(t, true, dtb.IsFailed)
		Contain(t, dtb.Logs.String(), logMSG)
	})
	t.Run("when io.Reader encounters an error", func(t *testing.T) {
		dtb := &doubles.TB{}
		subject := asserter(dtb)
		err := rnd.Error()
		_ = subject.ReadAll(iotest.ErrReader(err), msg...)
		Equal(t, true, dtb.IsFailed)
		Contain(t, dtb.Logs.String(), err.Error())
		Contain(t, dtb.Logs.String(), logMSG)
	})
}

func TestRegisterEqual(t *testing.T) {
	type SpecialT struct{ A, B int }

	assert.RegisterEqual[SpecialT](func(v1, v2 SpecialT) bool {
		return v1.A == v2.B && v1.B == v2.A
	})

	dtb := &doubles.TB{}
	sandbox.Run(func() {
		assert.Equal(dtb,
			SpecialT{A: 1, B: 2},
			SpecialT{A: 2, B: 1})
	})
	assert.False(t, dtb.IsFailed)
}

package random_test

import (
	"math/rand"
	"testing"
	"time"

	"github.com/adamluzsi/testcase"
	"github.com/adamluzsi/testcase/assert"
	"github.com/adamluzsi/testcase/random"
)

func TestMake(t *testing.T) {
	type ExampleStruct struct {
		String string
		Int    int
	}
	type Example struct {
		Bool          bool
		String        string
		Int           int
		Int8          int8
		Int16         int16
		Int32         int32
		Int64         int64
		UIntPtr       uintptr
		UInt          uint
		UInt8         uint8
		UInt16        uint16
		UInt32        uint32
		UInt64        uint64
		Float32       float32
		Float64       float64
		ArrayOfString [1]string
		ArrayOfInt    [1]int
		SliceOfString []string
		SliceOfInt    []int
		ChanOfString  chan string
		ChanOfInt     chan int
		Map           map[string]int
		StringPtr     *string
		IntPtr        *int
		Func          func()
		Duration      time.Duration
		Time          time.Time
		ExampleStruct
	}

	s := testcase.NewSpec(t)
	s.NoSideEffect()

	rnd := testcase.Let(s, func(t *testcase.T) *random.Random {
		return random.New(rand.NewSource(time.Now().UnixNano()))
	})

	s.Test("bool", func(t *testcase.T) {
		t.Eventually(func(it assert.It) {
			it.Must.True(random.Make[bool](rnd.Get(t)))
		})

		t.Eventually(func(it assert.It) {
			it.Must.False(random.Make[bool](rnd.Get(t)))
		})
	})
	s.Test("string", func(t *testcase.T) {
		str := random.Make[string](rnd.Get(t))
		t.Eventually(func(it assert.It) {
			it.Must.NotEmpty(str)

			it.Must.NotEqual(
				random.Make[string](rnd.Get(t)),
				random.Make[string](rnd.Get(t)),
			)
		})
	})
	s.Test("Integer", func(t *testcase.T) {
		t.Eventually(func(it assert.It) {
			it.Must.NotEqual(random.Make[int](rnd.Get(t)), int(0))
			it.Must.NotEqual(random.Make[int8](rnd.Get(t)), int8(0))
			it.Must.NotEqual(random.Make[int16](rnd.Get(t)), int16(0))
			it.Must.NotEqual(random.Make[int32](rnd.Get(t)), int32(0))
			it.Must.NotEqual(random.Make[int64](rnd.Get(t)), int64(0))
		})
	})
	s.Test("unsigned Integer", func(t *testcase.T) {
		t.Eventually(func(it assert.It) {
			it.Must.NotEqual(random.Make[uint](rnd.Get(t)), uint(0))
			it.Must.NotEqual(random.Make[uint8](rnd.Get(t)), uint8(0))
			it.Must.NotEqual(random.Make[uint16](rnd.Get(t)), uint16(0))
			it.Must.NotEqual(random.Make[uint32](rnd.Get(t)), uint32(0))
			it.Must.NotEqual(random.Make[uint64](rnd.Get(t)), uint64(0))
		})
	})
	s.Test("uintptr", func(t *testcase.T) {
		t.Eventually(func(it assert.It) {
			it.Must.NotEqual(random.Make[uintptr], uintptr(0))
		})
	})
	s.Test("floating point number", func(t *testcase.T) {
		t.Eventually(func(it assert.It) {
			it.Must.NotEqual(random.Make[float64](rnd.Get(t)), float64(0))
			it.Must.NotEqual(random.Make[float32](rnd.Get(t)), float32(0))
		})
	})
	s.Test("array", func(t *testcase.T) {
		t.Eventually(func(it assert.It) {
			var strs [42]string = random.Make[[42]string](rnd.Get(t))
			it.Must.NotNil(strs)

			it.Must.AnyOf(func(anyOf *assert.AnyOf) {
				for _, str := range strs {
					anyOf.Test(func(it assert.It) {
						it.Must.NotEmpty(str)
					})
				}
			})
		})

		t.Eventually(func(it assert.It) {
			var ints [42]int = random.Make[[42]int](rnd.Get(t))
			it.Must.NotNil(ints)

			it.Must.AnyOf(func(anyOf *assert.AnyOf) {
				for _, str := range ints {
					anyOf.Test(func(it assert.It) {
						it.Must.NotEmpty(str)
					})
				}
			})
		})
	})
	s.Test("slice", func(t *testcase.T) {
		t.Eventually(func(it assert.It) {
			var strs []string = random.Make[[]string](rnd.Get(t))
			it.Must.NotNil(strs)

			it.Must.AnyOf(func(anyOf *assert.AnyOf) {
				for _, str := range strs {
					anyOf.Test(func(it assert.It) {
						it.Must.NotEmpty(str)
					})
				}
			})
		})

		t.Eventually(func(it assert.It) {
			var ints []int = random.Make[[]int](rnd.Get(t))
			it.Must.NotNil(ints)

			it.Must.AnyOf(func(anyOf *assert.AnyOf) {
				for _, str := range ints {
					anyOf.Test(func(it assert.It) {
						it.Must.NotEmpty(str)
					})
				}
			})
		})
	})
	s.Test("chan", func(t *testcase.T) {
		t.Eventually(func(it assert.It) {
			ch := random.Make[chan int](rnd.Get(t))
			it.Must.NotNil(ch)
			it.Log("should be still empty")
			go func() { ch <- 42 }()
			it.Must.Equal(42, <-ch)
		})
	})
	s.Test("map", func(t *testcase.T) {
		t.Eventually(func(it assert.It) {
			m := random.Make[map[string]int](rnd.Get(t))
			it.Must.NotNil(m)
			it.Must.NotEmpty(m)

			for k, v := range m {
				it.Must.NotEmpty(k)
				it.Must.NotEmpty(v)
			}
		})
	})
	s.Test("pointer", func(t *testcase.T) {
		t.Eventually(func(it assert.It) {
			m := random.Make[*int](rnd.Get(t))
			it.Must.NotNil(m)
			it.Must.NotEmpty(*m)
		})
	})

	s.Test("func", func(t *testcase.T) {
		t.Log("there is no reasonable way to return a random function value")
		t.Must.Nil(random.Make[func()](rnd.Get(t)))
	})

	s.Test(`duration`, func(t *testcase.T) {
		t.Eventually(func(it assert.It) {
			it.Must.NotEmpty(random.Make[time.Duration](rnd.Get(t)))
		})
	})

	s.Test(`time`, func(t *testcase.T) {
		t.Eventually(func(it assert.It) {
			tm := random.Make[time.Time](rnd.Get(t))
			it.Must.False(tm.IsZero())
			it.Must.NotEqual(
				random.Make[time.Time](rnd.Get(t)),
				random.Make[time.Time](rnd.Get(t)),
			)
		})
	})

	s.Test("struct", func(t *testcase.T) {
		makeExample := func() Example {
			return random.Make[Example](rnd.Get(t))
		}
		t.Eventually(func(it assert.It) {
			it.Must.True(makeExample().Bool)
		})
		v := makeExample()
		t.Eventually(func(it assert.It) { it.Must.NotEmpty(makeExample().String) })
		t.Eventually(func(it assert.It) { it.Must.NotEmpty(makeExample().Int) })
		t.Eventually(func(it assert.It) { it.Must.NotEmpty(makeExample().Int8) })
		t.Eventually(func(it assert.It) { it.Must.NotEmpty(makeExample().Int16) })
		t.Eventually(func(it assert.It) { it.Must.NotEmpty(makeExample().Int32) })
		t.Eventually(func(it assert.It) { it.Must.NotEmpty(makeExample().Int64) })
		t.Eventually(func(it assert.It) { it.Must.NotEmpty(makeExample().UIntPtr) })
		t.Eventually(func(it assert.It) { it.Must.NotEmpty(makeExample().UInt) })
		t.Eventually(func(it assert.It) { it.Must.NotEmpty(makeExample().UInt8) })
		t.Eventually(func(it assert.It) { it.Must.NotEmpty(makeExample().UInt16) })
		t.Eventually(func(it assert.It) { it.Must.NotEmpty(makeExample().UInt32) })
		t.Eventually(func(it assert.It) { it.Must.NotEmpty(makeExample().UInt64) })
		t.Eventually(func(it assert.It) { it.Must.NotEmpty(makeExample().Float32) })
		t.Eventually(func(it assert.It) { it.Must.NotEmpty(makeExample().Float64) })
		t.Eventually(func(it assert.It) { it.Must.NotNil(v.ArrayOfInt) })
		t.Eventually(func(it assert.It) { it.Must.NotNil(v.ArrayOfString) })
		t.Eventually(func(it assert.It) { it.Must.NotNil(v.SliceOfInt) })
		t.Eventually(func(it assert.It) { it.Must.NotNil(v.SliceOfString) })
		t.Eventually(func(it assert.It) { it.Must.NotNil(v.ChanOfInt) })
		t.Eventually(func(it assert.It) { it.Must.NotNil(v.ChanOfString) })
		t.Eventually(func(it assert.It) { it.Must.NotNil(v.Map) })
		t.Eventually(func(it assert.It) { it.Must.NotNil(*v.StringPtr) })
		t.Eventually(func(it assert.It) { it.Must.NotNil(*v.IntPtr) })
		t.Eventually(func(it assert.It) { it.Must.NotEmpty(v.ExampleStruct.Int) })
		t.Eventually(func(it assert.It) { it.Must.NotEmpty(v.ExampleStruct.String) })
		t.Eventually(func(it assert.It) { it.Must.Nil(v.Func) })
		t.Eventually(func(it assert.It) { it.Must.NotEqual(time.Duration(0), v.Duration) })
		t.Eventually(func(it assert.It) { it.Must.False(v.Time.IsZero()) })
	})
	s.Test("*struct", func(t *testcase.T) {
		makeExample := func() *Example {
			return rnd.Get(t).Make(new(Example)).(*Example)
		}
		t.Eventually(func(it assert.It) {
			it.Must.True(makeExample().Bool)
		})
		v := makeExample()
		t.Eventually(func(it assert.It) { it.Must.NotEmpty(makeExample().String) })
		t.Eventually(func(it assert.It) { it.Must.NotEmpty(makeExample().Int) })
		t.Eventually(func(it assert.It) { it.Must.NotEmpty(makeExample().Int8) })
		t.Eventually(func(it assert.It) { it.Must.NotEmpty(makeExample().Int16) })
		t.Eventually(func(it assert.It) { it.Must.NotEmpty(makeExample().Int32) })
		t.Eventually(func(it assert.It) { it.Must.NotEmpty(makeExample().Int64) })
		t.Eventually(func(it assert.It) { it.Must.NotEmpty(makeExample().UIntPtr) })
		t.Eventually(func(it assert.It) { it.Must.NotEmpty(makeExample().UInt) })
		t.Eventually(func(it assert.It) { it.Must.NotEmpty(makeExample().UInt8) })
		t.Eventually(func(it assert.It) { it.Must.NotEmpty(makeExample().UInt16) })
		t.Eventually(func(it assert.It) { it.Must.NotEmpty(makeExample().UInt32) })
		t.Eventually(func(it assert.It) { it.Must.NotEmpty(makeExample().UInt64) })
		t.Eventually(func(it assert.It) { it.Must.NotEmpty(makeExample().Float32) })
		t.Eventually(func(it assert.It) { it.Must.NotEmpty(makeExample().Float64) })
		t.Eventually(func(it assert.It) { it.Must.NotNil(v.ArrayOfInt) })
		t.Eventually(func(it assert.It) { it.Must.NotNil(v.ArrayOfString) })
		t.Eventually(func(it assert.It) { it.Must.NotNil(v.SliceOfInt) })
		t.Eventually(func(it assert.It) { it.Must.NotNil(v.SliceOfString) })
		t.Eventually(func(it assert.It) { it.Must.NotNil(v.ChanOfInt) })
		t.Eventually(func(it assert.It) { it.Must.NotNil(v.ChanOfString) })
		t.Eventually(func(it assert.It) { it.Must.NotNil(v.Map) })
		t.Eventually(func(it assert.It) { it.Must.NotNil(*v.StringPtr) })
		t.Eventually(func(it assert.It) { it.Must.NotNil(*v.IntPtr) })
		t.Eventually(func(it assert.It) { it.Must.NotEmpty(v.ExampleStruct.Int) })
		t.Eventually(func(it assert.It) { it.Must.NotEmpty(v.ExampleStruct.String) })
		t.Eventually(func(it assert.It) { it.Must.Nil(v.Func) })
		t.Eventually(func(it assert.It) { it.Must.NotEqual(time.Duration(0), v.Duration) })
		t.Eventually(func(it assert.It) { it.Must.False(v.Time.IsZero()) })
	})
}

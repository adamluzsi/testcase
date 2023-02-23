package random_test

import (
	"github.com/adamluzsi/testcase/pp"
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
			it.Must.True(rnd.Get(t).Make(bool(false)).(bool))
		})

		t.Eventually(func(it assert.It) {
			it.Must.False(rnd.Get(t).Make(bool(false)).(bool))
		})
	})
	s.Test("string", func(t *testcase.T) {
		str := rnd.Get(t).Make(string(""))
		t.Eventually(func(it assert.It) {
			it.Must.NotEmpty(str)

			it.Must.NotEqual(
				rnd.Get(t).Make(string("")),
				rnd.Get(t).Make(string("")),
			)
		})
	})
	s.Test("Integer", func(t *testcase.T) {
		t.Eventually(func(it assert.It) {
			it.Must.NotEqual(rnd.Get(t).Make(int(0)).(int), int(0))
			it.Must.NotEqual(rnd.Get(t).Make(int8(0)).(int8), int8(0))
			it.Must.NotEqual(rnd.Get(t).Make(int16(0)).(int16), int16(0))
			it.Must.NotEqual(rnd.Get(t).Make(int32(0)).(int32), int32(0))
			it.Must.NotEqual(rnd.Get(t).Make(int64(0)).(int64), int64(0))
		})
	})
	s.Test("unsigned Integer", func(t *testcase.T) {
		t.Eventually(func(it assert.It) {
			it.Must.NotEqual(rnd.Get(t).Make(uint(0)).(uint), uint(0))
			it.Must.NotEqual(rnd.Get(t).Make(uint8(0)).(uint8), uint8(0))
			it.Must.NotEqual(rnd.Get(t).Make(uint16(0)).(uint16), uint16(0))
			it.Must.NotEqual(rnd.Get(t).Make(uint32(0)).(uint32), uint32(0))
			it.Must.NotEqual(rnd.Get(t).Make(uint64(0)).(uint64), uint64(0))
		})
	})
	s.Test("uintptr", func(t *testcase.T) {
		t.Eventually(func(it assert.It) {
			it.Must.NotEqual(rnd.Get(t).Make(uintptr(0)), uintptr(0))
		})
	})
	s.Test("floating point number", func(t *testcase.T) {
		t.Eventually(func(it assert.It) {
			it.Must.NotEqual(rnd.Get(t).Make(float64(0)).(float64), float64(0))
			it.Must.NotEqual(rnd.Get(t).Make(float32(0)).(float32), float32(0))
		})
	})
	s.Test("array", func(t *testcase.T) {
		t.Eventually(func(it assert.It) {
			var strings [42]string = rnd.Get(t).Make([42]string{}).([42]string)
			it.Must.NotNil(strings)

			it.Must.AnyOf(func(anyOf *assert.AnyOf) {
				for _, str := range strings {
					anyOf.Test(func(it assert.It) {
						it.Must.NotEmpty(str)
					})
				}
			})
		})

		t.Eventually(func(it assert.It) {
			var ints [42]int = rnd.Get(t).Make([42]int{}).([42]int)
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
			var strings []string = rnd.Get(t).Make([]string{}).([]string)
			it.Must.NotNil(strings)

			it.Must.AnyOf(func(anyOf *assert.AnyOf) {
				for _, str := range strings {
					anyOf.Test(func(it assert.It) {
						it.Must.NotEmpty(str)
					})
				}
			})
		})

		t.Eventually(func(it assert.It) {
			var ints []int = rnd.Get(t).Make([]int{}).([]int)
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
			ch := rnd.Get(t).Make(make(chan int)).(chan int)
			it.Must.NotNil(ch)
			it.Log("should be still empty")
			go func() { ch <- 42 }()
			it.Must.Equal(42, <-ch)
		})
	})
	s.Test("map", func(t *testcase.T) {
		t.Eventually(func(it assert.It) {
			m := rnd.Get(t).Make(map[string]int{}).(map[string]int)
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
			m := rnd.Get(t).Make((*int)(nil)).(*int)
			it.Must.NotNil(m)
			it.Must.NotEmpty(*m)
		})
	})

	s.Test("func", func(t *testcase.T) {
		t.Log("there is no reasonable way to return a random function value")
		t.Must.Nil(rnd.Get(t).Make(Example{}).(Example).Func)
	})

	s.Test(`duration`, func(t *testcase.T) {
		t.Eventually(func(it assert.It) {
			it.Must.NotEmpty(rnd.Get(t).Make(time.Duration(0)).(time.Duration))
		})
	})

	s.Test(`time`, func(t *testcase.T) {
		t.Eventually(func(it assert.It) {
			tm := rnd.Get(t).Make(time.Time{}).(time.Time)
			it.Must.False(tm.IsZero())
			it.Must.NotEqual(
				rnd.Get(t).Make(time.Time{}).(time.Time),
				rnd.Get(t).Make(time.Time{}).(time.Time),
			)
		})
	})

	s.Test("struct", func(t *testcase.T) {
		makeExample := func() Example {
			return rnd.Get(t).Make(Example{}).(Example)
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

func TestSlice_smoke(t *testing.T) {
	it := assert.MakeIt(t)
	eventually := assert.EventuallyWithin(5 * time.Second)
	rnd := random.New(random.CryptoSeed{})
	length := rnd.IntB(1, 5)
	slice1 := random.Slice[int](length, rnd.Int)
	pp.PP(slice1)
	it.Must.Equal(length, len(slice1))
	it.Must.NotEmpty(slice1)
	it.Must.AnyOf(func(a *assert.AnyOf) {
		for _, v := range slice1 {
			a.Test(func(it assert.It) {
				it.Must.NotEmpty(v)
			})
		}
	})
	eventually.Assert(t, func(it assert.It) {
		slice2 := random.Slice[int](length, rnd.Int)
		it.Must.Equal(len(slice1), len(slice2))
		it.Must.NotEqual(slice1, slice2)
	})
}

func TestMap_smoke(t *testing.T) {
	it := assert.MakeIt(t)
	eventually := assert.EventuallyWithin(5 * time.Second)
	rnd := random.New(random.CryptoSeed{})
	length := rnd.IntB(1, 5)
	map1 := random.Map[string, int](length, func() (string, int) {
		return rnd.String(), rnd.Int()
	})
	it.Must.Equal(length, len(map1))
	it.Must.NotEmpty(map1)
	it.Must.AnyOf(func(a *assert.AnyOf) {
		for k, v := range map1 {
			a.Test(func(it assert.It) {
				it.Must.NotEmpty(k)
				it.Must.NotEmpty(v)
			})
		}
	})
	eventually.Assert(t, func(it assert.It) {
		map2 := random.Map[string, int](length, random.KV(rnd.String, rnd.Int))
		it.Must.Equal(len(map1), len(map2))
		it.Must.NotEqual(map1, map2)
	})
}

func TestMap_whenNotEnoughUniqueKeyCanBeGenerated_thenItReturnsWithLess(t *testing.T) {
	it := assert.MakeIt(t)
	rnd := random.New(random.CryptoSeed{})
	map1 := random.Map[string, int](10, func() (string, int) {
		keys := []string{"foo", "bar", "baz"}
		return rnd.SliceElement(keys).(string), rnd.Int()
	})
	it.Must.NotEmpty(map1)
	it.Must.AnyOf(func(a *assert.AnyOf) {
		for k, v := range map1 {
			a.Test(func(it assert.It) {
				it.Must.NotEmpty(k)
				it.Must.NotEmpty(v)
			})
		}
	})
}

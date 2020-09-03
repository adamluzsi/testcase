package fixtures

import (
	"reflect"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestNew(t *testing.T) {
	t.Run("fixtures.New", func(t *testing.T) {
		t.Run("when given value is a struct", func(t *testing.T) {
			SharedSpecAssertions(t, func() *Example {
				return New(Example{}).(*Example)
			})
		})

		t.Run("when given value is a pointer to a struct", func(t *testing.T) {
			SharedSpecAssertions(t, func() *Example {
				return New(&Example{}).(*Example)
			})
		})
	})
}

func SharedSpecAssertions(t *testing.T, subject func() *Example) {
	require.NotNil(t, subject())

	t.Run("bool", func(t *testing.T) {
		t.Parallel()

		bools := make(map[bool]struct{})

		for i := 0; i < 1024; i++ {
			bools[subject().Bool] = struct{}{}
		}

		if _, ok := bools[true]; !ok {
			t.Fail()
		}

		if _, ok := bools[false]; !ok {
			t.Fail()
		}

	})
	t.Run("string", func(t *testing.T) {
		t.Parallel()

		require.NotEmpty(t, subject().String)
	})
	t.Run("Integer", func(t *testing.T) {
		t.Parallel()

		require.NotEqual(t, subject().Int, 0)
		require.NotEqual(t, subject().Int8, 0)
		require.NotEqual(t, subject().Int16, 0)
		require.NotEqual(t, subject().Int32, 0)
		require.NotEqual(t, subject().Int64, 0)
	})
	t.Run("unsigned Integer", func(t *testing.T) {
		t.Parallel()

		require.NotEqual(t, subject().UInt, 0)
		require.NotEqual(t, subject().UInt8, 0)
		require.NotEqual(t, subject().UInt16, 0)
		require.NotEqual(t, subject().UInt32, 0)
		require.NotEqual(t, subject().UInt64, 0)
	})
	t.Run("uintptr", func(t *testing.T) {
		t.Parallel()

		require.NotEqual(t, subject().UIntPtr, 0)
	})
	t.Run("floating point number", func(t *testing.T) {
		t.Parallel()

		require.NotEqual(t, subject().Float32, 0)
		require.NotEqual(t, subject().Float64, 0)
	})
	t.Run("complex numbers", func(t *testing.T) {
		t.Parallel()

		require.NotEqual(t, subject().Complex64, 0)
		require.NotEqual(t, subject().Complex128, 0)
	})
	t.Run("array", func(t *testing.T) {
		t.Parallel()

		require.NotNil(t, subject().ArrayOfInt)
		require.NotNil(t, subject().ArrayOfString)
	})
	t.Run("slice", func(t *testing.T) {
		t.Parallel()

		require.NotNil(t, subject().SliceOfInt)
		require.NotNil(t, subject().SliceOfString)
	})
	t.Run("chan", func(t *testing.T) {
		t.Parallel()

		require.NotNil(t, subject().ChanOfInt)
		require.NotNil(t, subject().ChanOfString)
	})
	t.Run("map", func(t *testing.T) {
		t.Parallel()

		require.NotNil(t, subject().Map)
	})
	t.Run("pointer", func(t *testing.T) {
		t.Parallel()

		require.NotNil(t, *subject().StringPtr)
		require.NotNil(t, *subject().IntPtr)
	})
	t.Run("struct", func(t *testing.T) {
		t.Parallel()

		require.NotEqual(t, subject().ExampleStruct.Int, 0)
		require.NotEmpty(t, subject().ExampleStruct.String)
	})
	t.Run("func", func(t *testing.T) {
		t.Parallel()

		require.Nil(t, subject().Func)
	})
	t.Run(`duration`, func(t *testing.T) {
		t.Parallel()

		require.NotEqual(t, time.Duration(0), subject().Duration)
	})
	t.Run(`time`, func(t *testing.T) {
		t.Parallel()

		require.NotEqual(t, time.Time{}, subject().Time)
	})
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
	Complex64     complex64
	Complex128    complex128
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

type ExampleStruct struct {
	String string
	Int    int
}

//----------------------------------------------------- reflect ------------------------------------------------------//

func TestBaseValueOf(t *testing.T) {
	subject := func(input interface{}) reflect.Value {
		return baseValueOf(input)
	}

	SpecForPrimitiveNames(t, func(obj interface{}) string {
		return subject(obj).Type().Name()
	})

	type StructObject struct{}

	expectedValue := reflect.ValueOf(StructObject{})
	expectedValueType := expectedValue.Type()

	plainStruct := StructObject{}
	ptrToStruct := &plainStruct
	ptrToPtr := &ptrToStruct

	require.Equal(t, expectedValueType, subject(plainStruct).Type())
	require.Equal(t, expectedValueType, subject(ptrToStruct).Type())
	require.Equal(t, expectedValueType, subject(ptrToPtr).Type())
}

func TestBaseTypeOf(t *testing.T) {
	subject := func(obj interface{}) reflect.Type {
		return baseTypeOf(obj)
	}

	SpecForPrimitiveNames(t, func(obj interface{}) string {
		return subject(obj).Name()
	})

	type StructObject struct{}

	expectedValueType := reflect.TypeOf(StructObject{})

	plainStruct := StructObject{}
	ptrToStruct := &plainStruct
	ptrToPtr := &ptrToStruct

	require.Equal(t, expectedValueType, subject(plainStruct))
	require.Equal(t, expectedValueType, subject(ptrToStruct))
	require.Equal(t, expectedValueType, subject(ptrToPtr))
}

func SpecForPrimitiveNames(spec *testing.T, subject func(entity interface{}) string) {
	spec.Run("when given object is a bool", func(t *testing.T) {
		t.Parallel()

		require.Equal(t, "bool", subject(true))
	})

	spec.Run("when given object is a string", func(t *testing.T) {
		t.Parallel()

		require.Equal(t, "string", subject(`42`))
	})

	spec.Run("when given object is a int", func(t *testing.T) {
		t.Parallel()

		require.Equal(t, "int", subject(int(42)))
	})

	spec.Run("when given object is a int8", func(t *testing.T) {
		t.Parallel()

		require.Equal(t, "int8", subject(int8(42)))
	})

	spec.Run("when given object is a int16", func(t *testing.T) {
		t.Parallel()

		require.Equal(t, "int16", subject(int16(42)))
	})

	spec.Run("when given object is a int32", func(t *testing.T) {
		t.Parallel()

		require.Equal(t, "int32", subject(int32(42)))
	})

	spec.Run("when given object is a int64", func(t *testing.T) {
		t.Parallel()

		require.Equal(t, "int64", subject(int64(42)))
	})

	spec.Run("when given object is a uintptr", func(t *testing.T) {
		t.Parallel()

		require.Equal(t, "uintptr", subject(uintptr(42)))
	})

	spec.Run("when given object is a uint", func(t *testing.T) {
		t.Parallel()

		require.Equal(t, "uint", subject(uint(42)))
	})

	spec.Run("when given object is a uint8", func(t *testing.T) {
		t.Parallel()

		require.Equal(t, "uint8", subject(uint8(42)))
	})

	spec.Run("when given object is a uint16", func(t *testing.T) {
		t.Parallel()

		require.Equal(t, "uint16", subject(uint16(42)))
	})

	spec.Run("when given object is a uint32", func(t *testing.T) {
		t.Parallel()

		require.Equal(t, "uint32", subject(uint32(42)))
	})

	spec.Run("when given object is a uint64", func(t *testing.T) {
		t.Parallel()

		require.Equal(t, "uint64", subject(uint64(42)))
	})

	spec.Run("when given object is a float32", func(t *testing.T) {
		t.Parallel()

		require.Equal(t, "float32", subject(float32(42)))
	})

	spec.Run("when given object is a float64", func(t *testing.T) {
		t.Parallel()

		require.Equal(t, "float64", subject(float64(42)))
	})

	spec.Run("when given object is a complex64", func(t *testing.T) {
		t.Parallel()

		require.Equal(t, "complex64", subject(complex64(42)))
	})

	spec.Run("when given object is a complex128", func(t *testing.T) {
		t.Parallel()

		require.Equal(t, "complex128", subject(complex128(42)))
	})
}

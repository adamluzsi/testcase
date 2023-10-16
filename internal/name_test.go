package internal_test

import (
	"testing"

	"go.llib.dev/testcase/assert"
	"go.llib.dev/testcase/internal"
)

func TestSymbolicName(t *testing.T) {
	subject := internal.SymbolicName

	t.Run("when given object is a bool", func(t *testing.T) {
		t.Parallel()

		assert.Must(t).Equal("bool", subject(true))
	})

	t.Run("when given object is a string", func(t *testing.T) {
		t.Parallel()

		assert.Must(t).Equal("string", subject(`42`))
	})

	t.Run("when given object is a int", func(t *testing.T) {
		t.Parallel()

		assert.Must(t).Equal("int", subject(int(42)))
	})

	t.Run("when given object is a int8", func(t *testing.T) {
		t.Parallel()

		assert.Must(t).Equal("int8", subject(int8(42)))
	})

	t.Run("when given object is a int16", func(t *testing.T) {
		t.Parallel()

		assert.Must(t).Equal("int16", subject(int16(42)))
	})

	t.Run("when given object is a int32", func(t *testing.T) {
		t.Parallel()

		assert.Must(t).Equal("int32", subject(int32(42)))
	})

	t.Run("when given object is a int64", func(t *testing.T) {
		t.Parallel()

		assert.Must(t).Equal("int64", subject(int64(42)))
	})

	t.Run("when given object is a uintptr", func(t *testing.T) {
		t.Parallel()

		assert.Must(t).Equal("uintptr", subject(uintptr(42)))
	})

	t.Run("when given object is a uint", func(t *testing.T) {
		t.Parallel()

		assert.Must(t).Equal("uint", subject(uint(42)))
	})

	t.Run("when given object is a uint8", func(t *testing.T) {
		t.Parallel()

		assert.Must(t).Equal("uint8", subject(uint8(42)))
	})

	t.Run("when given object is a uint16", func(t *testing.T) {
		t.Parallel()

		assert.Must(t).Equal("uint16", subject(uint16(42)))
	})

	t.Run("when given object is a uint32", func(t *testing.T) {
		t.Parallel()

		assert.Must(t).Equal("uint32", subject(uint32(42)))
	})

	t.Run("when given object is a uint64", func(t *testing.T) {
		t.Parallel()

		assert.Must(t).Equal("uint64", subject(uint64(42)))
	})

	t.Run("when given object is a float32", func(t *testing.T) {
		t.Parallel()

		assert.Must(t).Equal("float32", subject(float32(42)))
	})

	t.Run("when given object is a float64", func(t *testing.T) {
		t.Parallel()

		assert.Must(t).Equal("float64", subject(float64(42)))
	})

	t.Run("when given object is a complex64", func(t *testing.T) {
		t.Parallel()

		assert.Must(t).Equal("complex64", subject(complex64(42)))
	})

	t.Run("when given object is a complex128", func(t *testing.T) {
		t.Parallel()

		assert.Must(t).Equal("complex128", subject(complex128(42)))
	})

	type TestSymbolicNameStruct struct{}
	const expectedStructName = `internal_test.TestSymbolicNameStruct`

	t.Run("when given struct is from different package than the current one", func(t *testing.T) {
		t.Parallel()

		o := TestSymbolicNameStruct{}
		assert.Must(t).Equal(expectedStructName, subject(o))
	})

	t.Run("when given object is a pointer of a struct", func(t *testing.T) {
		t.Parallel()

		assert.Must(t).Equal(expectedStructName, subject(&TestSymbolicNameStruct{}))
	})

	t.Run("when given object is a pointer of a pointer of a struct", func(t *testing.T) {
		t.Parallel()

		o := &TestSymbolicNameStruct{}

		assert.Must(t).Equal(expectedStructName, subject(&o))
	})
}

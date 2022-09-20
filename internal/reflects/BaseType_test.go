package reflects_test

import (
	"github.com/adamluzsi/testcase/assert"
	"github.com/adamluzsi/testcase/internal/fixtures/mypkg"
	"github.com/adamluzsi/testcase/internal/reflects"
	"reflect"
	"testing"
)

func TestBaseTypeOf(t *testing.T) {
	subject := func(obj interface{}) reflect.Type {
		return reflects.BaseTypeOf(obj)
	}

	SpecForPrimitiveNames(t, func(obj interface{}) string {
		return subject(obj).Name()
	})

	expectedValueType := reflect.TypeOf(mypkg.ExampleStruct{})

	plainStruct := mypkg.ExampleStruct{}
	ptrToStruct := &plainStruct
	ptrToPtr := &ptrToStruct

	assert.Must(t).Equal(expectedValueType, subject(plainStruct))
	assert.Must(t).Equal(expectedValueType, subject(ptrToStruct))
	assert.Must(t).Equal(expectedValueType, subject(ptrToPtr))
}

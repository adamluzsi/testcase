package faultinject_test

import (
	"context"
	"testing"

	"github.com/adamluzsi/testcase/assert"
	"github.com/adamluzsi/testcase/faultinject"
)

func TestInject_smoke(t *testing.T) {
	assert.NotPanic(t, func() {
		type Tag1 struct{}
		faultinject.Inject(context.Background(), Tag1{})
	})
	assert.Panic(t, func() {
		faultinject.Inject(context.Background(), "non-struct-type")
	})
	assert.Panic(t, func() {
		faultinject.Inject(context.Background(), nil)
	})
}

func TestInject_onEmptyTagList(t *testing.T) {
	inCTX := context.Background()
	outCTX := faultinject.Inject(inCTX)
	assert.Equal(t, inCTX, outCTX)
}

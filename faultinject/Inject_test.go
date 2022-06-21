package faultinject_test

import (
	"context"
	"testing"

	"github.com/adamluzsi/testcase/assert"
	"github.com/adamluzsi/testcase/faultinject"
)

func TestInject_smoke(t *testing.T) {
	t.Cleanup(faultinject.Enable())

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
	t.Cleanup(faultinject.Enable())
	inCTX := context.Background()
	outCTX := faultinject.Inject(inCTX)
	assert.Equal(t, inCTX, outCTX)
}

func TestInject_onEnabledFalse(t *testing.T) {
	assert.False(t, faultinject.Enabled())
	inCTX := context.Background()
	outCTX := faultinject.Inject(inCTX, Tag1{})
	assert.Equal(t, inCTX, outCTX)
}

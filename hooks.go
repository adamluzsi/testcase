package testcase

import (
	"runtime"
	"testing"

	"github.com/adamluzsi/testcase/internal/caller"
)

const hookWarning = `you cannot create spec hooks after you used describe/when/and/then,
unless you create a new spec with the previously mentioned calls`

type hookBlock func(*T) func()

type hook struct {
	Block hookBlock
	Frame runtime.Frame
}

type hookOnce struct {
	Block func() func()
	Frame runtime.Frame
}

// Before give you the ability to run a block before each test case.
// This is ideal for doing clean ahead before each test case.
// The received *testing.T object is the same as the Test block *testing.T object
// This hook applied to this scope and anything that is nested from here.
// All setup block is stackable.
func (spec *Spec) Before(beforeBlock tBlock) {
	spec.testingTB.Helper()
	spec.Around(func(t *T) func() {
		beforeBlock(t)
		return func() {}
	})
}

// After give you the ability to run a block after each test case.
// This is ideal for running cleanups.
// The received *testing.T object is the same as the Then block *testing.T object
// This hook applied to this scope and anything that is nested from here.
// All setup block is stackable.
//
// DEPRECATED: use Spec.Before with T.Cleanup or Spec.Before with T.Defer instead
func (spec *Spec) After(afterBlock tBlock) {
	spec.testingTB.Helper()
	spec.Around(func(t *T) func() {
		return func() { afterBlock(t) }
	})
}

// Around give you the ability to create "Before" setup for each test case,
// with the additional ability that the returned function will be deferred to run after the Then block is done.
// This is ideal for setting up mocks, and then return the assertion request calls in the return func.
// This hook applied to this scope and anything that is nested from here.
// All setup block is stackable.
//
// DEPRECATED: use Spec.Before with T.Cleanup or Spec.Before with T.Defer instead
func (spec *Spec) Around(block hookBlock) {
	//fmt.Println(internal.GetFrame())
	spec.testingTB.Helper()
	if spec.immutable {
		spec.testingTB.Fatal(hookWarning)
	}
	frame, _ := caller.GetFrame()
	spec.hooks.Around = append(spec.hooks.Around, hook{
		Block: block,
		Frame: frame,
	})
}

// BeforeAll give you the ability to create a hook
// that runs only once before the test cases.
func (spec *Spec) BeforeAll(blk func(tb testing.TB)) {
	spec.testingTB.Helper()
	spec.AroundAll(func(tb testing.TB) func() {
		blk(tb)
		return func() {}
	})
}

// AfterAll give you the ability to create a hook
// that runs only once after all the test cases already ran.
//
// DEPRECATED: use Spec.BeforeAll with testing.TB#Cleanup
func (spec *Spec) AfterAll(blk func(tb testing.TB)) {
	spec.testingTB.Helper()
	spec.AroundAll(func(tb testing.TB) func() {
		return func() { blk(tb) }
	})
}

// AroundAll give you the ability to create a hook
// that first run before all test,
// then the returned lambda will run after the test cases.
//
// DEPRECATED: use Spec.BeforeAll with testing.TB#Cleanup
func (spec *Spec) AroundAll(blk func(tb testing.TB) func()) {
	spec.testingTB.Helper()
	if spec.immutable {
		spec.testingTB.Fatal(hookWarning)
	}
	frame, _ := caller.GetFrame()
	spec.hooks.AroundAll = append(spec.hooks.AroundAll, hookOnce{
		Block: func() func() { return blk(spec.testingTB) },
		Frame: frame,
	})
}

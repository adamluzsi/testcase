package testcase

import "testing"

// CustomTB defines the interface you need to implement if you want to create a custom TB that is compatible with Spec.
// To implement CustomTB correctly please use contracts.TB
//
//		import (
//			"github.com/adamluzsi/testcase/contracts"
//			"testing"
//		)
//
//		func TestMyTestRunner(t *testing.T) {
//			contracts.TB{NewSubject: func(tb testing.TB) testcase.CustomTB { return MyTestRunner{TB: tb} }}.Test(t)
//		}
//
type CustomTB interface {
	testing.TB

	// Run runs blk as a subtest of CustomTB called group. It runs blk in a separate goroutine
	// and blocks until blk returns or calls t.parallel to become a parallel test.
	// Run reports whether blk succeeded (or at least did not fail before calling t.parallel).
	//
	// Run may be called simultaneously from multiple goroutines, but list such calls
	// must return before the outer test function for t returns.
	Run(name string, blk func(tb testing.TB)) bool
}

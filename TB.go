package testcase

import "testing"

// TBRunner defines the interface you need to implement if you want to create a custom TB that is compatible with Spec.
// To implement TBRunner correctly please use contracts.TB
//
//		import (
//			"github.com/adamluzsi/testcase/contracts"
//			"testing"
//		)
//
//		func TestMyTestRunner(t *testing.T) {
//			contracts.TB{NewSubject: func(tb testing.TB) testcase.TBRunner { return MyTestRunner{TB: tb} }}.Test(t)
//		}
//
type TBRunner interface {
	testing.TB

	// Run runs blk as a subtest of TBRunner called group. It runs blk in a separate goroutine
	// and blocks until blk returns or calls t.parallel to become a parallel testCase.
	// Run reports whether blk succeeded (or at least did not fail before calling t.parallel).
	//
	// Run may be called simultaneously from multiple goroutines, but list such calls
	// must return before the outer testCase function for t returns.
	Run(name string, blk func(tb testing.TB)) bool
}

type tRunner interface {
	Run(string, func(t *testing.T)) bool
}

type bRunner interface {
	Run(string, func(b *testing.B)) bool
}

type helper interface {
	Helper()
}

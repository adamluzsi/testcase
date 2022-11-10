package testcase

import (
	"testing"

	"github.com/adamluzsi/testcase/internal/doubles"
)

// TBRunner defines the interface you need to implement if you want to create a custom TB that is compatible with Spec.
// To implement TBRunner correctly please use contracts.TB
//
//		import (
//			"github.com/adamluzsi/testcase/contracts"
//			"testing"
//		)
//
//		func TestMyTestRunner(t *testing.T) {
//			contracts.TB{Subject: func(tb testing.TB) testcase.TBRunner { return MyTestRunner{TB: tb} }}.Test(t)
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

type testingT interface {
	testing.TB
	tRunner
}

type tRunner interface {
	Run(string, func(t *testing.T)) bool
}

type testingB interface {
	testing.TB
	bRunner
}

type bRunner interface {
	Run(string, func(b *testing.B)) bool
}

type helper interface {
	Helper()
}

type anyTB interface {
	*T | *testing.T | *testing.B | *doubles.TB | *testing.TB | *TBRunner
}

type anyTBOrSpec interface {
	anyTB | *Spec
}

func ToSpec[TBS anyTBOrSpec](tbs TBS) *Spec {
	switch tbs := any(tbs).(type) {
	case *Spec:
		return tbs
	case *T:
		return NewSpec(tbs)
	case *testing.T:
		return NewSpec(tbs)
	case *testing.B:
		return NewSpec(tbs)
	case *doubles.TB:
		return NewSpec(tbs)
	case *testing.TB:
		return NewSpec(*tbs)
	default:
		panic("not implemented")
	}
}

func ToT[TBs anyTB](tb TBs) *T {
	switch tbs := (any)(tb).(type) {
	case *T:
		return tbs
	case *testing.T:
		return NewT(tbs, NewSpec(tbs))
	case *testing.B:
		return NewT(tbs, NewSpec(tbs))
	case *doubles.TB:
		return NewT(tbs, NewSpec(tbs))
	case *testing.TB:
		return NewT(*tbs, NewSpec(*tbs))
	case *TBRunner:
		return NewT(*tbs, NewSpec(*tbs))
	default:
		panic("not implemented")
	}
}

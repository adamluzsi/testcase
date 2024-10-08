package testcase

import (
	"reflect"
	"testing"

	"go.llib.dev/testcase/assert"
	"go.llib.dev/testcase/internal/doubles"
)

// TBRunner defines the interface you need to implement if you want to create a custom TB that is compatible with Spec.
// To implement TBRunner correctly please use contracts.TB
//
//	import (
//		"go.llib.dev/testcase/contracts"
//		"testing"
//	)
//
//	func TestMyTestRunner(t *testing.T) {
//		contracts.TB{Subject: func(tb testing.TB) testcase.TBRunner { return MyTestRunner{TB: tb} }}.Test(t)
//	}
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

type testingHelper interface {
	Helper()
}

type anyTB interface {
	*T | *testing.T | *testing.B | *testing.F | *doubles.TB | *testing.TB | *TBRunner | assert.It
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
	return toT(tb)
}

func toT(tb any) *T {
	switch tbs := (any)(tb).(type) {
	case *T:
		return tbs
	case *testing.T:
		return NewT(tbs)
	case *testing.B:
		return NewT(tbs)
	case *doubles.TB:
		return NewT(tbs)
	case *testing.TB:
		return toT(*tbs)
	case assert.It:
		return toT(tbs.TB)
	case testing.TB:
		return NewT(unwrapTestingTB(tbs))
	case *TBRunner:
		return toT(*tbs)
	default:
		panic("not implemented")
	}
}

var reflectTypeTestingTB = reflect.TypeOf((*testing.TB)(nil)).Elem()

func unwrapTestingTB(tb testing.TB) testing.TB {
	rtb := reflect.ValueOf(tb)
	if rtb.Kind() == reflect.Pointer {
		rtb = rtb.Elem()
	}
	if rtb.Kind() != reflect.Struct {
		return tb
	}

	rtbType := rtb.Type()
	NumField := rtbType.NumField()
	for i := 0; i < NumField; i++ {
		fieldType := rtbType.Field(i)
		// Implementing testing.TB is only possible when a struct includes an embedded field from the testing package.
		// This requirement arises because testing.TB has a private function method expectation that can only be implemented within the testing package scope.
		// As a result, we can identify the embedded field regardless of whether it is *testing.T, *testing.B, *testing.F, testing.TB, etc
		// by checking if the field itself is an embedded field and implements testing.TB.
		if fieldType.Anonymous && fieldType.Type.Implements(reflectTypeTestingTB) {
			testingTB, ok := rtb.Field(i).Interface().(testing.TB)
			if ok && testingTB != nil {
				return testingTB
			}
		}
	}
	return tb
}

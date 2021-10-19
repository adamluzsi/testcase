package assert

import "testing"

func Should(tb testing.TB) Asserter {
	return Asserter{
		Helper: tb.Helper,
		FailFn: tb.Error,
	}
}

func Must(tb testing.TB) Asserter {
	return Asserter{
		Helper: tb.Helper,
		FailFn: tb.Fatal,
	}
}

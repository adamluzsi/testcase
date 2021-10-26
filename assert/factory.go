package assert

import "testing"

type Factory struct {
	TB testing.TB
}

func (f Factory) Must() Asserter {
	return Asserter{
		TB:     f.TB,
		FailFn: f.TB.Fatal,
	}
}

func (f Factory) Should() Asserter {
	return Asserter{
		TB:     f.TB,
		FailFn: f.TB.Error,
	}
}

func Should(tb testing.TB) Asserter {
	return Factory{TB: tb}.Should()
}

func Must(tb testing.TB) Asserter {
	return Factory{TB: tb}.Must()
}

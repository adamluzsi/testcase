package teardown_test

import (
	"go.llib.dev/testcase/internal/teardown"
)

func offsetHelper(td *teardown.Teardown, fn interface{}, args ...interface{}) { td.Defer(fn, args...) }

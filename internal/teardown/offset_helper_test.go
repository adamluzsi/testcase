package teardown_test

import (
	"github.com/adamluzsi/testcase/internal/teardown"
)

func offsetHelper(td *teardown.Teardown, fn interface{}, args ...interface{}) { td.Defer(fn, args...) }

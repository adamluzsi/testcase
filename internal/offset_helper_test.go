package internal_test

import "github.com/adamluzsi/testcase/internal"

func offsetHelper(td *internal.Teardown, fn interface{}, args ...interface{}) { td.Defer(fn, args...) }

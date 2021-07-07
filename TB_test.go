package testcase

import "testing"

var (
	_ testingT = &testing.T{}
	_ testingB = &testing.B{}
)

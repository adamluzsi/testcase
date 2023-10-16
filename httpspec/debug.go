package httpspec

import "go.llib.dev/testcase"

var debug = testcase.Var[bool]{
	ID:   `httpspec:debug`,
	Init: func(t *testcase.T) bool { return false },
}

func Debug(s *testcase.Spec) {
	debug.LetValue(s, true)
}

func isDebugEnabled(t *testcase.T) bool {
	return debug.Get(t)
}

package httpspec

import "github.com/adamluzsi/testcase"

const debugLetVar = letVarPrefix + `debug`

type debugFlag struct{}

func setupDebug(s *testcase.Spec) {
	s.Let(debugLetVar, func(t *testcase.T) interface{} { return nil })
}

func Debug(s *testcase.Spec) {
	s.Let(debugLetVar, func(t *testcase.T) interface{} { return debugFlag{} })
}

func isDebugEnabled(t *testcase.T) bool {
	_, ok := t.I(debugLetVar).(debugFlag)
	return ok
}

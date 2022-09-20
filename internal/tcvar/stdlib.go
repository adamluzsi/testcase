package tcvar

import (
	"context"
	"github.com/adamluzsi/testcase"
)

func LetContext(s *testcase.Spec) testcase.Var[context.Context] {
	return testcase.Let(s, func(t *testcase.T) context.Context {
		return context.Background()
	})
}

func LetError(s *testcase.Spec) testcase.Var[error] {
	return testcase.Let(s, func(t *testcase.T) error {
		return t.Random.Error()
	})
}

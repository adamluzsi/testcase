package faultinject_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/adamluzsi/testcase"
	"github.com/adamluzsi/testcase/faultinject"
)

func TestFault(t *testing.T) {
	s := testcase.NewSpec(t)

	var receiver = testcase.Let(s, func(t *testcase.T) *ExampleReceiver { return &ExampleReceiver{} })

	var (
		packageV  = testcase.Let[string](s, nil)
		receiverV = testcase.Let[string](s, nil)
		functionV = testcase.Let[string](s, nil)
	)
	act := func(t *testcase.T) error {
		ctx := faultinject.Inject(context.Background(), faultinject.CallerFault{
			Package:  packageV.Get(t),
			Receiver: receiverV.Get(t),
			Function: functionV.Get(t),
		}, exampleErr.Get(t))
		return receiver.Get(t).Main(ctx)
	}

	s.Before(func(t *testcase.T) {
		faultinject.EnableForTest(t)
	})

	s.When("no identifier is given", func(s *testcase.Spec) {
		packageV.LetValue(s, "faultinject_test")
		receiverV.LetValue(s, "")
		functionV.LetValue(s, "")

		s.Then("it will inject error on matching package", func(t *testcase.T) {
			t.Must.ErrorIs(exampleErr.Get(t), act(t))
			t.Must.True(receiver.Get(t).MainRanFaultPoint)
			t.Must.False(receiver.Get(t).MainIsFinished)
		})
	})

	s.When("package is specified", func(s *testcase.Spec) {
		receiverV.LetValue(s, "")
		functionV.LetValue(s, "")

		s.And("it is matching with the callers", func(s *testcase.Spec) {
			packageV.LetValue(s, "faultinject_test")

			s.Then("it will inject error on first occasion for matching package", func(t *testcase.T) {
				t.Must.ErrorIs(exampleErr.Get(t), act(t))
				t.Must.True(receiver.Get(t).MainRanFaultPoint)
				t.Must.False(receiver.Get(t).MainIsFinished)
			})
		})

		s.And("it is not matching with the callers package", func(s *testcase.Spec) {
			packageV.LetValue(s, "injectfault_test")

			s.Then("error won't be injected on check", func(t *testcase.T) {
				t.Must.Nil(act(t))
				t.Must.True(receiver.Get(t).MainIsFinished)
			})
		})
	})

	s.When("receiver is specified", func(s *testcase.Spec) {
		packageV.LetValue(s, "")
		functionV.LetValue(s, "")

		s.And("it is matching with the callers", func(s *testcase.Spec) {
			receiverV.LetValue(s, "*ExampleReceiver")

			s.Then("it will inject error on first occasion for matching package", func(t *testcase.T) {
				t.Must.ErrorIs(exampleErr.Get(t), act(t))
				t.Must.True(receiver.Get(t).MainRanFaultPoint)
				t.Must.False(receiver.Get(t).MainIsFinished)
			})
		})

		s.And("it is not matching with the callers", func(s *testcase.Spec) {
			receiverV.LetValue(s, "*OtherReceiver")

			s.Then("error won't be injected on check", func(t *testcase.T) {
				t.Must.Nil(act(t))
				t.Must.True(receiver.Get(t).MainIsFinished)
			})
		})
	})

	s.When("function is specified", func(s *testcase.Spec) {
		packageV.LetValue(s, "")
		receiverV.LetValue(s, "")
		functionV.LetValue(s, "")

		s.And("it is matching with the callers", func(s *testcase.Spec) {
			functionV.LetValue(s, "Main")

			s.Then("it will inject error on the given function", func(t *testcase.T) {
				t.Must.ErrorIs(exampleErr.Get(t), act(t))
				t.Must.True(receiver.Get(t).MainRanFaultPoint)
				t.Must.False(receiver.Get(t).MainIsFinished)
			})

			for _, fnName := range []string{
				"OnErr",
				"OnValue",
			} {
				fnName := fnName
				s.And(fmt.Sprintf("it is a specific function call down in the stack (%s)", fnName), func(s *testcase.Spec) {
					functionV.LetValue(s, fnName)

					s.Then("it will inject error on the given function", func(t *testcase.T) {
						t.Must.ErrorIs(exampleErr.Get(t), act(t))
						t.Must.False(receiver.Get(t).MainRanFaultPoint)
					})
				})
			}
		})

		s.And("it is not matching with the callers", func(s *testcase.Spec) {
			functionV.LetValue(s, "OtherFunction")

			s.Then("error won't be injected on check", func(t *testcase.T) {
				t.Must.Nil(act(t))
				t.Must.True(receiver.Get(t).MainIsFinished)
			})
		})
	})
}

type ExampleReceiver struct {
	MainRanFaultPoint bool
	MainIsFinished    bool
}

func (r *ExampleReceiver) Main(ctx context.Context) error {
	if err := ctx.Err(); err != nil {
		r.MainRanFaultPoint = true
		return err
	}
	if err := r.OnErr(ctx); err != nil {
		return err
	}
	if err := r.OnValue(ctx); err != nil {
		return err
	}
	r.MainIsFinished = true
	return nil
}

func (r *ExampleReceiver) OnErr(ctx context.Context) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	return nil
}

func (r *ExampleReceiver) OnValue(ctx context.Context) error {
	type SomeTag struct{}
	if err, ok := ctx.Value(SomeTag{}).(error); ok {
		return err
	}
	return nil
}

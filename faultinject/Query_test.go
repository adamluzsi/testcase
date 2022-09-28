package faultinject_test

import (
	"github.com/adamluzsi/testcase"
	"github.com/adamluzsi/testcase/faultinject"
	"github.com/adamluzsi/testcase/internal/fixtures/mypkg"
	"github.com/adamluzsi/testcase/internal/fixtures/othpkg"
	"github.com/adamluzsi/testcase/internal/tcvar"
	"testing"
)

func TestCallerQuery(t *testing.T) {
	s := testcase.NewSpec(t)

	var (
		ctx        = tcvar.LetContext(s)
		exampleErr = tcvar.LetError(s)
		receiver   = testcase.Let(s, func(t *testcase.T) *mypkg.ExampleStruct { return &mypkg.ExampleStruct{} })
		query      = testcase.Let(s, func(t *testcase.T) faultinject.Query { return faultinject.Query{} })
	)
	act := func(t *testcase.T) error {
		faultinject.EnableForTest(t)
		return receiver.Get(t).Main(
			faultinject.Inject(ctx.Get(t), query.Get(t), exampleErr.Get(t)))
	}

	s.Then("it will inject error", func(t *testcase.T) {
		t.Must.ErrorIs(exampleErr.Get(t), act(t))
		t.Must.True(receiver.Get(t).MainRanFaultPoint)
		t.Must.False(receiver.Get(t).MainIsFinished)
	})

	s.When("package is specified with the a value", func(s *testcase.Spec) {
		value := testcase.Let[any](s, nil)

		query.Let(s, func(t *testcase.T) faultinject.Query {
			return query.Super(t).PackageOf(value.Get(t))
		})

		s.And("it match the receiver's stack", func(s *testcase.Spec) {
			value.LetValue(s, mypkg.ExampleStruct{})

			s.Then("it will inject error on first occasion for matching package", func(t *testcase.T) {
				t.Must.ErrorIs(exampleErr.Get(t), act(t))
				t.Must.True(receiver.Get(t).MainRanFaultPoint)
				t.Must.False(receiver.Get(t).MainIsFinished)
			})
		})

		s.And("it does not match the receiver's stack", func(s *testcase.Spec) {
			value.LetValue(s, othpkg.ExampleStruct{})

			s.Then("error won't be injected on check", func(t *testcase.T) {
				t.Must.Nil(act(t))
				t.Must.True(receiver.Get(t).MainIsFinished)
			})
		})
	})

	s.When("package is specified with a symbolic name as string", func(s *testcase.Spec) {
		value := testcase.Let[string](s, nil)

		query.Let(s, func(t *testcase.T) faultinject.Query {
			return query.Super(t).Package(value.Get(t))
		})

		s.And("it match the receiver's stack", func(s *testcase.Spec) {
			value.LetValue(s, "mypkg")

			s.Then("it will inject error on first occasion for matching package", func(t *testcase.T) {
				t.Must.ErrorIs(exampleErr.Get(t), act(t))
				t.Must.True(receiver.Get(t).MainRanFaultPoint)
				t.Must.False(receiver.Get(t).MainIsFinished)
			})
		})

		s.And("it does not match the receiver's stack", func(s *testcase.Spec) {
			value.LetValue(s, "othpkg")

			s.Then("error won't be injected on check", func(t *testcase.T) {
				t.Must.Nil(act(t))
				t.Must.True(receiver.Get(t).MainIsFinished)
			})
		})
	})

	s.When("receiver is specified with an example value", func(s *testcase.Spec) {
		value := testcase.Let[any](s, nil)

		query.Let(s, func(t *testcase.T) faultinject.Query {
			return query.Super(t).Receiver(value.Get(t))
		})

		s.And("it match the receiver's stack", func(s *testcase.Spec) {
			value.LetValue(s, mypkg.ExampleStruct{})

			s.Then("it will inject error on first occasion for matching package", func(t *testcase.T) {
				t.Must.ErrorIs(exampleErr.Get(t), act(t))
				t.Must.True(receiver.Get(t).MainRanFaultPoint)
				t.Must.False(receiver.Get(t).MainIsFinished)
			})
		})

		s.And("it match the receiver's stack with a pointer type value", func(s *testcase.Spec) {
			value.Let(s, func(t *testcase.T) any {
				return &mypkg.ExampleStruct{}
			})

			s.Then("it will inject error on first occasion for matching package", func(t *testcase.T) {
				t.Must.ErrorIs(exampleErr.Get(t), act(t))
				t.Must.True(receiver.Get(t).MainRanFaultPoint)
				t.Must.False(receiver.Get(t).MainIsFinished)
			})
		})

		s.And("it does not match the receiver's stack", func(s *testcase.Spec) {
			value.LetValue(s, othpkg.ExampleStruct{})

			s.Then("error won't be injected on check", func(t *testcase.T) {
				t.Must.Nil(act(t))
				t.Must.True(receiver.Get(t).MainIsFinished)
			})
		})
	})

	s.When("function is specified with the a value", func(s *testcase.Spec) {
		value := testcase.Let[any](s, nil)

		query.Let(s, func(t *testcase.T) faultinject.Query {
			return query.Super(t).Function(value.Get(t))
		})

		s.And("it match the receiver's stack", func(s *testcase.Spec) {
			value.Let(s, func(t *testcase.T) any {
				return (*mypkg.ExampleStruct)(nil).Main
			})

			s.Then("it will inject error on first occasion for matching package", func(t *testcase.T) {
				t.Must.ErrorIs(exampleErr.Get(t), act(t))
				t.Must.True(receiver.Get(t).MainRanFaultPoint)
				t.Must.False(receiver.Get(t).MainIsFinished)
			})
		})

		s.And("it does not match the receiver's stack", func(s *testcase.Spec) {
			value.Let(s, func(t *testcase.T) any {
				return othpkg.ExampleStruct.Foo
			})

			s.Then("error won't be injected on check", func(t *testcase.T) {
				t.Must.Nil(act(t))
				t.Must.True(receiver.Get(t).MainIsFinished)
			})
		})
	})
}

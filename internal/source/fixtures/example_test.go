package fixtures_test

import (
	"context"
	"testing"

	tc "github.com/adamluzsi/testcase"
	"github.com/adamluzsi/testcase/internal/source/fixtures"
	"github.com/adamluzsi/testcase/random"
)

var ctxV = tc.Var[context.Context]{
	ID: "caller context",
	Init: func(t *tc.T) context.Context {
		return context.Background()
	},
}

func TestExample(t *testing.T) {
	s := tc.NewSpec(t)

	v := tc.Let(s, func(t *tc.T) int {
		return t.Random.Int()
	})
	myType := tc.Let(s, func(t *tc.T) *fixtures.MyType {
		return &fixtures.MyType{
			IntField: v.Get(t),
		}
	})

	var (
		foo = func() {}
		bar = func() {}
		baz = func() {}
		qux = func() {}
	)

	s.Before(func(t *tc.T) { bar() })
	s.Before(beforeHelper)

	s.BeforeAll(func(testing.TB) { foo() })
	s.BeforeAll(beforeAllHelper)

	subject := func(t *tc.T) error {
		return myType.Get(t).MyFunc(ctxV.Get(t))
	}

	s.Context("spec context", func(s *tc.Spec) {
		s.Before(func(t *tc.T) {
			baz()
		})
		s.Before(func(t *tc.T) {
			if t.Random.Bool() {
				qux()
			}
		})

		s.Test("test1", func(t *tc.T) {
			t.Must.Nil(subject(t))
		})

		s.Test("test2", SpecExampleTest2)
	})
}

func SpecExampleTest2(t *tc.T) {
	t.Log("OK")
}

var someRandomGlobalString string

func beforeAllHelper(tb testing.TB) {
	someRandomGlobalString = random.New(random.CryptoSeed{}).String()
}

func beforeHelper(t *tc.T) {
	ctxV.Set(t, context.WithValue(ctxV.Get(t), "foo", "bar"))
}

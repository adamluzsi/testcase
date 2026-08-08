package sandbox_test

import (
	"fmt"
	"runtime"
	"testing"

	"go.llib.dev/testcase"
	"go.llib.dev/testcase/assert"
	"go.llib.dev/testcase/let"
	"go.llib.dev/testcase/sandbox"
)

func TestRun(t *testing.T) {
	s := testcase.NewSpec(t)

	var (
		fn = testcase.Let[func()](s, nil)
	)
	act := func(t *testcase.T) sandbox.O {
		return sandbox.Run(fn.Get(t))
	}

	s.When("the sandboxed function runs without an issue", func(s *testcase.Spec) {
		fn.Let(s, func(t *testcase.T) func() {
			return func() {}
		})

		s.Then("runs without an issue", func(t *testcase.T) {
			outcome := act(t)
			assert.Must(t).True(outcome.OK)
			assert.Must(t).Nil(outcome.PanicValue)
			assert.Must(t).False(outcome.Goexit)
		})
	})

	s.When("the sandboxed function panics", func(s *testcase.Spec) {
		expectedPanicValue := testcase.Let(s, func(t *testcase.T) string {
			return t.Random.String()
		})
		fn.Let(s, func(t *testcase.T) func() {
			return func() {
				panic(expectedPanicValue.Get(t))
			}
		})

		s.Then("it reports the panic value", func(t *testcase.T) {
			outcome := act(t)
			assert.Must(t).False(outcome.OK)
			assert.Must(t).False(outcome.Goexit)
			assert.Must(t).Equal(any(expectedPanicValue.Get(t)), outcome.PanicValue)
		})

		s.Then("it returns the panic stack trace", func(t *testcase.T) {
			outcome := act(t)
			assert.Must(t).False(outcome.OK)
			assert.Must(t).False(outcome.Goexit)
			assert.Must(t).Equal(outcome.Trace(), outcome.Trace())
			assert.Must(t).Contains(outcome.Trace(), fmt.Sprintf("panic: %v", expectedPanicValue.Get(t)))
			_, file, _, _ := runtime.Caller(0)
			assert.Must(t).Contains(outcome.Trace(), file)
		})
	})

	s.When("the sandboxed function calls runtime.Goexit", func(s *testcase.Spec) {
		fn.Let(s, func(t *testcase.T) func() {
			return func() { runtime.Goexit() }
		})

		s.Then("it reports the Goexit", func(t *testcase.T) {
			outcome := act(t)
			assert.Must(t).False(outcome.OK)
			assert.Must(t).True(outcome.Goexit)
		})
	})
}

func TestRunOutcome_OnNotOK(t *testing.T) {
	t.Run("happy", func(t *testing.T) {
		var ran bool
		sandbox.Run(func() {}).
			OnNotOK(func() { ran = true })
		assert.False(t, ran)
	})
	t.Run("rainy", func(t *testing.T) {
		var ran bool
		sandbox.Run(func() { panic("boom") }).
			OnNotOK(func() { ran = true })
		assert.True(t, ran)
	})
	t.Run("rainy plus nil block", func(t *testing.T) {
		sandbox.Run(func() { panic("boom") }).
			OnNotOK(nil)
	})
}

func TestRun_frameless(t *testing.T) {
	s := testcase.NewSpec(t)

	var (
		fn = let.Var[func()](s, nil)
	)
	act := let.Act(func(t *testcase.T) sandbox.O {
		return sandbox.Run(fn.Get(t))
	})

	s.Context("OK", func(s *testcase.Spec) {
		fn.Let(s, func(t *testcase.T) func() {
			return func() {}
		})

		s.Test("successful execution is reported", func(t *testcase.T) {
			o := act(t)

			assert.True(t, o.OK)
			assert.False(t, o.Goexit)
			assert.False(t, o.Panic)
			assert.Nil(t, o.PanicValue)
		})
	})

	s.Context("Goexit", func(s *testcase.Spec) {
		fn.Let(s, func(t *testcase.T) func() {
			return func() {
				runtime.Goexit()
			}
		})

		s.Test("non OK execution is reported with Goexit mentioned", func(t *testcase.T) {
			o := act(t)

			assert.False(t, o.OK)
			assert.True(t, o.Goexit)
			assert.False(t, o.Panic)
			assert.Nil(t, o.PanicValue)
		})
	})

	s.Context("panic", func(s *testcase.Spec) {
		val := let.Var[any](s, nil)
		fn.Let(s, func(t *testcase.T) func() {
			return func() {
				panic(val.Get(t))
			}
		})

		s.Context("nil panic value", func(s *testcase.Spec) {
			val.LetValue(s, nil)

			s.Test("panic with nil value is reported", func(t *testcase.T) {
				o := act(t)

				assert.False(t, o.OK)
				assert.False(t, o.Goexit)
				assert.True(t, o.Panic)
				assert.Nil(t, o.PanicValue)
			})
		})

		s.Context("non-nil panic value", func(s *testcase.Spec) {
			exp := let.Error(s)

			val.Let(s, func(t *testcase.T) any {
				return exp.Get(t)
			})

			s.Test("panic with the expected panic value is reported", func(t *testcase.T) {
				o := act(t)

				assert.False(t, o.OK)
				assert.False(t, o.Goexit)
				assert.True(t, o.Panic)
				assert.Equal[any](t, o.PanicValue, exp.Get(t))
			})
		})
	})
}

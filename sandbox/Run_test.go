package sandbox_test

import (
	"fmt"
	"runtime"
	"testing"

	"go.llib.dev/testcase"
	"go.llib.dev/testcase/assert"
	"go.llib.dev/testcase/sandbox"
)

func TestRun(t *testing.T) {
	s := testcase.NewSpec(t)

	var (
		fn = testcase.Let[func()](s, nil)
	)
	act := func(t *testcase.T) sandbox.RunOutcome {
		return sandbox.Run(fn.Get(t))
	}

	s.When("the sandboxed function runs without an issue", func(s *testcase.Spec) {
		fn.Let(s, func(t *testcase.T) func() {
			return func() {}
		})

		s.Then("runs without an issue", func(t *testcase.T) {
			outcome := act(t)
			t.Must.True(outcome.OK)
			t.Must.Nil(outcome.PanicValue)
			t.Must.False(outcome.Goexit)
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
			t.Must.False(outcome.OK)
			t.Must.False(outcome.Goexit)
			t.Must.Equal(any(expectedPanicValue.Get(t)), outcome.PanicValue)
		})

		s.Then("it returns the panic stack trace", func(t *testcase.T) {
			outcome := act(t)
			t.Must.False(outcome.OK)
			t.Must.False(outcome.Goexit)
			t.Must.Equal(outcome.Trace(), outcome.Trace())
			t.Must.Contains(outcome.Trace(), fmt.Sprintf("panic: %v", expectedPanicValue.Get(t)))
			_, file, _, _ := runtime.Caller(0)
			t.Must.Contains(outcome.Trace(), file)
		})
	})

	s.When("the sandboxed function calls runtime.Goexit", func(s *testcase.Spec) {
		fn.Let(s, func(t *testcase.T) func() {
			return func() { runtime.Goexit() }
		})

		s.Then("it reports the Goexit", func(t *testcase.T) {
			outcome := act(t)
			t.Must.False(outcome.OK)
			t.Must.True(outcome.Goexit)
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

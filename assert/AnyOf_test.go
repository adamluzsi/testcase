package assert_test

import (
	"go.llib.dev/testcase/random"
	"go.llib.dev/testcase/sandbox"
	"testing"

	"go.llib.dev/testcase"
	"go.llib.dev/testcase/assert"
	"go.llib.dev/testcase/internal/doubles"
)

func TestA(t *testing.T) {
	s := testcase.NewSpec(t)

	stub := testcase.Let(s, func(t *testcase.T) *doubles.TB {
		return &doubles.TB{}
	})
	anyOf := testcase.Let(s, func(t *testcase.T) *assert.A {
		return &assert.A{TB: stub.Get(t), Fail: stub.Get(t).Fail}
	})
	subject := func(t *testcase.T, blk func(it assert.It)) {
		anyOf.Get(t).Case(blk)
	}

	s.When(`there is at least one .Case with non failing ran`, func(s *testcase.Spec) {
		s.Before(func(t *testcase.T) {
			subject(t, func(it assert.It) { /* no fail */ })
		})

		s.Then(`AnyOf yields no failure on .Finish`, func(t *testcase.T) {
			anyOf.Get(t).Finish()
			t.Must.Equal(false, stub.Get(t).IsFailed)
		})

		s.Then("AnyOf.OK will be true, because one of the test passed", func(t *testcase.T) {
			anyOf.Get(t).Finish()

			t.Must.True(anyOf.Get(t).OK())
		})

		s.And(`and new .Case calls are made`, func(s *testcase.Spec) {
			additionalTestBlkRan := testcase.LetValue(s, false)
			s.Before(func(t *testcase.T) {
				subject(t, func(it assert.It) { additionalTestBlkRan.Set(t, true) })
			})

			s.Then("AnyOf.OK will be true, because one of the test passed", func(t *testcase.T) {
				anyOf.Get(t).Finish()

				t.Must.True(anyOf.Get(t).OK())
			})

			s.Then(`AnyOf yields no failure on .Finish`, func(t *testcase.T) {
				anyOf.Get(t).Finish()
				t.Must.Equal(false, stub.Get(t).IsFailed)
			})

			s.Then(`AnyOf will skip running additional test blocks`, func(t *testcase.T) {
				anyOf.Get(t).Finish()

				t.Must.Equal(false, additionalTestBlkRan.Get(t))
			})
		})
	})

	s.When(`.Case fails with .FailNow`, func(s *testcase.Spec) {
		s.Before(func(t *testcase.T) {
			subject(t, func(it assert.It) { it.Must.True(false) })
		})

		s.Then(`AnyOf yields failure on .Finish`, func(t *testcase.T) {
			anyOf.Get(t).Finish()
			t.Must.True(stub.Get(t).IsFailed)
		})

		s.Then("AnyOf.OK will yield false due to no passing test", func(t *testcase.T) {
			anyOf.Get(t).Finish()

			t.Must.False(anyOf.Get(t).OK())
		})

		s.And(`but there is one as well that pass`, func(s *testcase.Spec) {
			s.Before(func(t *testcase.T) {
				subject(t, func(it assert.It) {})
			})

			s.Then(`AnyOf yields no failure on .Finish`, func(t *testcase.T) {
				anyOf.Get(t).Finish()
				t.Must.Equal(false, stub.Get(t).IsFailed)
			})

			s.Then("AnyOf.OK will be true, because one of the test passed", func(t *testcase.T) {
				anyOf.Get(t).Finish()

				t.Must.True(anyOf.Get(t).OK())
			})
		})
	})
}

func TestA_Case_cleanup(t *testing.T) {
	h := assert.Must(t)
	stub := &doubles.TB{}
	anyOf := &assert.A{
		TB:   stub,
		Fail: stub.Fail,
	}

	var cleanupRan bool
	anyOf.Case(func(it assert.It) {
		it.Must.TB.Cleanup(func() { cleanupRan = true })
		it.Must.True(false) // fail it
	})
	h.True(cleanupRan, "cleanup should have ran already after leaving the block of AnyOf.Case")

	anyOf.Finish()
	h.True(stub.IsFailed, "the provided testing.TB should have failed")
}

func TestAnyOf_Test_race(t *testing.T) {
	stub := &doubles.TB{}
	anyOf := &assert.A{
		TB:   stub,
		Fail: stub.Fail,
	}
	testcase.Race(func() {
		anyOf.Case(func(it assert.It) {})
	}, func() {
		anyOf.Case(func(it assert.It) {})
	}, func() {
		anyOf.Finish()
	})
}

func TestOneOf(t *testing.T) {
	s := testcase.NewSpec(t)

	stub := testcase.Let(s, func(t *testcase.T) *doubles.TB {
		return &doubles.TB{}
	})
	vs := testcase.Let(s, func(t *testcase.T) []string {
		return random.Slice(t.Random.IntBetween(3, 7), func() string {
			return t.Random.String()
		})
	})

	const msg = "optional assertion explanation"
	blk := testcase.LetValue[func(assert.It, string)](s, nil)
	act := func(t *testcase.T) sandbox.RunOutcome {
		return sandbox.Run(func() {
			assert.OneOf(stub.Get(t), vs.Get(t), blk.Get(t), msg)
		})
	}

	s.When("passed block has no issue", func(s *testcase.Spec) {
		blk.Let(s, func(t *testcase.T) func(assert.It, string) {
			return func(it assert.It, s string) {}
		})

		s.Then("testing.TB is OK", func(t *testcase.T) {
			act(t)

			t.Must.False(stub.Get(t).IsFailed)
		})

		s.Then("execution context is not killed", func(t *testcase.T) {
			t.Must.True(act(t).OK)
		})

		s.Then("assert message explanation is not logged", func(t *testcase.T) {
			act(t)

			t.Must.NotContain(stub.Get(t).Logs.String(), msg)
		})
	})

	s.When("passed keeps failing with testing.TB#FailNow", func(s *testcase.Spec) {
		blk.Let(s, func(t *testcase.T) func(assert.It, string) {
			return func(it assert.It, s string) { it.FailNow() }
		})

		s.Then("testing.TB is failed", func(t *testcase.T) {
			act(t)

			t.Must.True(stub.Get(t).IsFailed)
		})

		s.Then("execution context is interrupted with FailNow", func(t *testcase.T) {
			out := act(t)
			t.Must.False(out.OK)
			t.Must.True(out.Goexit)
		})

		s.Then("assert message explanation is logged using the testing.TB", func(t *testcase.T) {
			act(t)

			t.Must.Contain(stub.Get(t).Logs.String(), msg)
		})

		s.Then("assertion failure message includes the assertion helper name", func(t *testcase.T) {
			act(t)

			t.Must.Contain(stub.Get(t).Logs.String(), "OneOf")
			t.Must.Contain(stub.Get(t).Logs.String(), "None of the element matched the expectations")
		})
	})

	s.When("assertion pass only for one of the slice element", func(s *testcase.Spec) {
		blk.Let(s, func(t *testcase.T) func(assert.It, string) {
			expected := t.Random.SliceElement(vs.Get(t)).(string)
			return func(it assert.It, got string) {
				it.Must.Equal(expected, got)
			}
		})

		s.Then("testing.TB is OK", func(t *testcase.T) {
			act(t)

			t.Must.False(stub.Get(t).IsFailed)
		})

		s.Then("execution context is not killed", func(t *testcase.T) {
			t.Must.True(act(t).OK)
		})

		s.Then("assert message explanation is not logged", func(t *testcase.T) {
			act(t)

			t.Must.NotContain(stub.Get(t).Logs.String(), msg)
		})
	})
}

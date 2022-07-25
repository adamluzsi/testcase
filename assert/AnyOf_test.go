package assert_test

import (
	"testing"

	"github.com/adamluzsi/testcase"
	"github.com/adamluzsi/testcase/assert"
	"github.com/adamluzsi/testcase/internal/doubles"
)

func TestAnyOf(t *testing.T) {
	s := testcase.NewSpec(t)

	stub := testcase.Let(s, func(t *testcase.T) *doubles.TB {
		return &doubles.TB{}
	})
	anyOf := testcase.Let(s, func(t *testcase.T) *assert.AnyOf {
		return &assert.AnyOf{TB: stub.Get(t), Fail: stub.Get(t).Fail}
	})
	subject := func(t *testcase.T, blk func(it assert.It)) {
		anyOf.Get(t).Test(blk)
	}

	s.When(`there is at least one .Test with non failing ran`, func(s *testcase.Spec) {
		s.Before(func(t *testcase.T) {
			subject(t, func(it assert.It) { /* no fail */ })
		})

		s.Then(`AnyOf yields no failure on .Finish`, func(t *testcase.T) {
			anyOf.Get(t).Finish()
			t.Must.Equal(false, stub.Get(t).IsFailed)
		})

		s.And(`and new .Test calls are made`, func(s *testcase.Spec) {
			additionalTestBlkRan := testcase.LetValue(s, false)
			s.Before(func(t *testcase.T) {
				subject(t, func(it assert.It) { additionalTestBlkRan.Set(t, true) })
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

	s.When(`.Test fails with .FailNow`, func(s *testcase.Spec) {
		s.Before(func(t *testcase.T) {
			subject(t, func(it assert.It) { it.Must.True(false) })
		})

		s.Then(`AnyOf yields failure on .Finish`, func(t *testcase.T) {
			anyOf.Get(t).Finish()
			t.Must.True(true, stub.Get(t).IsFailed)
		})

		s.And(`but there is one as well that pass`, func(s *testcase.Spec) {
			s.Before(func(t *testcase.T) {
				subject(t, func(it assert.It) {})
			})

			s.Then(`AnyOf yields no failure on .Finish`, func(t *testcase.T) {
				anyOf.Get(t).Finish()
				t.Must.Equal(false, stub.Get(t).IsFailed)
			})
		})
	})
}

func TestAnyOf_Test_cleanup(t *testing.T) {
	h := assert.Must(t)
	stub := &doubles.TB{}
	anyOf := &assert.AnyOf{
		TB:   stub,
		Fail: stub.Fail,
	}

	var cleanupRan bool
	anyOf.Test(func(it assert.It) {
		it.Must.TB.Cleanup(func() { cleanupRan = true })
		it.Must.True(false) // fail it
	})
	h.True(cleanupRan, "cleanup should have ran already after leaving the block of AnyOf.Test")

	anyOf.Finish()
	h.True(stub.IsFailed, "the provided testing.TB should have failed")
}

func TestAnyOf_Test_race(t *testing.T) {
	stub := &doubles.TB{}
	anyOf := &assert.AnyOf{
		TB:   stub,
		Fail: stub.Fail,
	}
	testcase.Race(func() {
		anyOf.Test(func(it assert.It) {})
	}, func() {
		anyOf.Test(func(it assert.It) {})
	}, func() {
		anyOf.Finish()
	})
}

//func TestAnyOf_smoke(t *testing.T) {
//	assert.Should(t).AnyOf(func(a *assert.AnyOf) {
//		//a.Test(func(it assert.It) {})
//		a.Test(func(it assert.It) {it.Must.True(false)})
//	})
//	t.Log(`after`)
//}

package assert_test

import (
	"context"
	"testing"
	"time"

	"go.llib.dev/testcase"
	"go.llib.dev/testcase/assert"
	"go.llib.dev/testcase/internal/doubles"
	"go.llib.dev/testcase/let"
	"go.llib.dev/testcase/random"
	"go.llib.dev/testcase/sandbox"
)

func TestEqual(t *testing.T) {
	s := testcase.NewSpec(t)

	var (
		tb = let.Var(s, func(t *testcase.T) *doubles.TB {
			return &doubles.TB{}
		})
		v = let.Var(s, func(t *testcase.T) any {
			return t.Random.String()
		})
		oth = let.Var(s, func(t *testcase.T) any {
			return v.Get(t)
		})
	)
	act := func(t *testcase.T) (out sandbox.O) {
		dtb, v, oth := tb.Get(t), v.Get(t), oth.Get(t)
		// bounded, so a comparison that blocks, e.g. while draining a channel, fails instead of hanging the suite.
		assert.Within(t, 5*time.Second, func(context.Context) {
			out = sandbox.Run(func() { assert.Equal(dtb, v, oth) })
		})
		return out
	}

	thenItPasses := func(s *testcase.Spec) {
		s.Then("it passes", func(t *testcase.T) {
			out := act(t)

			assert.False(t, tb.Get(t).IsFailed, assert.MessageF("%s", tb.Get(t).Logs.String()))
			assert.True(t, out.OK)
		})
	}
	thenItFails := func(s *testcase.Spec) {
		s.Then("it fails", func(t *testcase.T) {
			out := act(t)

			assert.True(t, tb.Get(t).IsFailed)
			assert.False(t, out.OK)
		})
	}

	thenItPasses(s)

	s.When("the values differ", func(s *testcase.Spec) {
		oth.Let(s, func(t *testcase.T) any {
			return random.Unique(func() any { return t.Random.String() }, v.Get(t))
		})

		thenItFails(s)
	})

	s.When("the values are structs with an unexported pointer field", func(s *testcase.Spec) {
		name := let.String(s)
		pointee := let.Int(s)
		v.Let(s, func(t *testcase.T) any {
			n := pointee.Get(t)
			return equalTestWithPtr{Name: name.Get(t), ptr: &n}
		})
		oth.Let(s, func(t *testcase.T) any {
			n := pointee.Get(t)
			return equalTestWithPtr{Name: name.Get(t), ptr: &n}
		})

		s.Then("it passes, as the pointed values are equal, even though the pointers differ", func(t *testcase.T) {
			out := act(t)

			assert.False(t, tb.Get(t).IsFailed, assert.MessageF("%s", tb.Get(t).Logs.String()))
			assert.True(t, out.OK)
		})

		s.And("the pointed values differ", func(s *testcase.Spec) {
			oth.Let(s, func(t *testcase.T) any {
				val := oth.Super(t).(equalTestWithPtr)
				n := *val.ptr + 1
				val.ptr = &n
				return val
			})

			thenItFails(s)
		})
	})

	s.When("the values are structs with an unexported interface field holding a non-comparable value", func(s *testcase.Spec) {
		texts := let.Var(s, func(t *testcase.T) []string {
			return random.Slice(t.Random.IntBetween(1, 3), func() string {
				return t.Random.String()
			})
		})
		v.Let(s, func(t *testcase.T) any {
			return equalTestWithIface{iface: append([]string(nil), texts.Get(t)...)}
		})
		oth.Let(s, func(t *testcase.T) any {
			return equalTestWithIface{iface: append([]string(nil), texts.Get(t)...)}
		})

		thenItPasses(s)

		s.And("the held values differ", func(s *testcase.Spec) {
			oth.Let(s, func(t *testcase.T) any {
				val := oth.Super(t).(equalTestWithIface)
				val.iface = append(val.iface.([]string), t.Random.String())
				return val
			})

			thenItFails(s)
		})
	})

	s.When("the values are non-comparable structs with an unexported struct field", func(s *testcase.Spec) {
		texts := let.Var(s, func(t *testcase.T) []string {
			return random.Slice(t.Random.IntBetween(1, 3), func() string {
				return t.Random.String()
			})
		})
		v.Let(s, func(t *testcase.T) any {
			return equalTestWithNested{nested: equalTestNested{texts: append([]string(nil), texts.Get(t)...)}}
		})
		oth.Let(s, func(t *testcase.T) any {
			return equalTestWithNested{nested: equalTestNested{texts: append([]string(nil), texts.Get(t)...)}}
		})

		thenItPasses(s)

		s.And("the nested struct values differ", func(s *testcase.Spec) {
			oth.Let(s, func(t *testcase.T) any {
				val := oth.Super(t).(equalTestWithNested)
				val.nested.texts = append(val.nested.texts, t.Random.String())
				return val
			})

			thenItFails(s)
		})
	})

	s.When("the values are structs with an unexported buffered channel field", func(s *testcase.Spec) {
		elems := let.Var(s, func(t *testcase.T) []int {
			return random.Slice(t.Random.IntBetween(1, equalTestChanCap), func() int {
				return t.Random.Int()
			})
		})
		// each call makes a distinct channel, buffering the same elements
		makeValue := func(t *testcase.T) equalTestWithChan {
			ch := make(chan int, equalTestChanCap)
			for _, e := range elems.Get(t) {
				ch <- e
			}
			return equalTestWithChan{ch: ch}
		}
		v.Let(s, func(t *testcase.T) any { return makeValue(t) })
		oth.Let(s, func(t *testcase.T) any { return makeValue(t) })

		s.Then("it passes, as the buffered elements are equal, even though the channels differ", func(t *testcase.T) {
			out := act(t)

			assert.False(t, tb.Get(t).IsFailed, assert.MessageF("%s", tb.Get(t).Logs.String()))
			assert.True(t, out.OK)
		})

		s.Then("the buffered elements are kept in their original order", func(t *testcase.T) {
			act(t)

			assert.Equal(t, elems.Get(t), drainEqualTestChan(v.Get(t).(equalTestWithChan).ch))
			assert.Equal(t, elems.Get(t), drainEqualTestChan(oth.Get(t).(equalTestWithChan).ch))
		})

		s.And("the buffered elements differ", func(s *testcase.Spec) {
			oth.Let(s, func(t *testcase.T) any {
				val := oth.Super(t).(equalTestWithChan)
				vs := drainEqualTestChan(val.ch)
				vs[len(vs)-1]++
				for _, e := range vs {
					val.ch <- e
				}
				return val
			})

			thenItFails(s)
		})

		s.And("one of the channels is nil", func(s *testcase.Spec) {
			oth.Let(s, func(t *testcase.T) any { return equalTestWithChan{} })

			thenItFails(s)
		})

		s.And("the channels are closed", func(s *testcase.Spec) {
			s.Before(func(t *testcase.T) {
				close(v.Get(t).(equalTestWithChan).ch)
				close(oth.Get(t).(equalTestWithChan).ch)
			})

			thenItPasses(s)

			s.Then("the buffered elements are kept", func(t *testcase.T) {
				act(t)

				assert.Equal(t, elems.Get(t), drainEqualTestChan(v.Get(t).(equalTestWithChan).ch))
				assert.Equal(t, elems.Get(t), drainEqualTestChan(oth.Get(t).(equalTestWithChan).ch))
			})
		})
	})

	s.When("the values are pointers to structs whose first field is a struct", func(s *testcase.Spec) {
		inner := let.Int(s)
		v.Let(s, func(t *testcase.T) any {
			return &equalTestOuter{Inner: equalTestInner{V: inner.Get(t)}}
		})
		oth.Let(s, func(t *testcase.T) any {
			return &equalTestOuter{Inner: equalTestInner{V: inner.Get(t)}}
		})

		thenItPasses(s)

		s.And("the first fields differ", func(s *testcase.Spec) {
			oth.Let(s, func(t *testcase.T) any {
				return &equalTestOuter{Inner: equalTestInner{V: inner.Get(t) + 1}}
			})

			thenItFails(s)
		})
	})

	s.When("the values reference themselves through an unexported field", func(s *testcase.Spec) {
		values := let.Var(s, func(t *testcase.T) []int {
			return random.Slice(t.Random.IntBetween(1, 5), func() int {
				return t.Random.Int()
			})
		})
		v.Let(s, func(t *testcase.T) any { return makeEqualTestRing(values.Get(t)) })
		oth.Let(s, func(t *testcase.T) any { return makeEqualTestRing(values.Get(t)) })

		thenItPasses(s)

		s.And("a value along the cycle differs", func(s *testcase.Spec) {
			oth.Let(s, func(t *testcase.T) any {
				vs := append([]int(nil), values.Get(t)...)
				vs[len(vs)-1]++
				return makeEqualTestRing(vs)
			})

			thenItFails(s)
		})
	})
}

type equalTestWithPtr struct {
	Name string
	ptr  *int
}

type equalTestWithIface struct{ iface any }

type equalTestWithNested struct{ nested equalTestNested }

// equalTestNested is non-comparable on purpose, so it can't be compared with the == operator.
type equalTestNested struct{ texts []string }

const equalTestChanCap = 3

type equalTestWithChan struct{ ch chan int }

type equalTestOuter struct {
	Inner equalTestInner
	Name  string
}

type equalTestInner struct{ V int }

type equalTestRing struct {
	V    int
	next *equalTestRing
}

func makeEqualTestRing(values []int) equalTestRing {
	nodes := make([]*equalTestRing, len(values))
	for i, v := range values {
		nodes[i] = &equalTestRing{V: v}
	}
	for i := range nodes {
		nodes[i].next = nodes[(i+1)%len(nodes)]
	}
	return *nodes[0]
}

func drainEqualTestChan(ch chan int) []int {
	var vs []int
	for n := len(ch); len(vs) < n; {
		vs = append(vs, <-ch)
	}
	return vs
}

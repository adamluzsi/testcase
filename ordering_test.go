package testcase

import (
	"testing"
	"time"

	"github.com/adamluzsi/testcase/assert"
	"github.com/adamluzsi/testcase/fixtures"
	"github.com/adamluzsi/testcase/internal"
)

var ord = Var[orderer]{ID: `orderer`}

func cpyOrdOut(src []int) []int {
	dst := make([]int, len(src))
	copy(dst, src)
	return dst
}

func cpyOrdInput(src []func()) []func() {
	dst := make([]func(), len(src))
	copy(dst, src)
	return dst
}

func genOrdInput(out *[]int) []func() {
	var fns []func()
	for i := 0; i < 42; i++ {
		n := i // copy with pass by value
		fns = append(fns, func() { *out = append(*out, n) })
	}
	return fns
}

func runOrdInput(fns []func(), out *[]int) []int {
	*out = []int{}
	for _, fn := range fns {
		fn()
	}
	return cpyOrdOut(*out)
}

////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

func TestNullOrderer_Order(t *testing.T) {
	s := NewSpec(t)
	s.NoSideEffect()

	ord.Let(s, func(t *T) orderer {
		return nullOrderer{}
	})

	s.Describe(`Order`, func(s *Spec) {
		subject := func(t *T, input []func()) {
			ord.Get(t).Order(input)
		}

		s.Test(`.Order should not affect the order of the id list`, func(t *T) {
			out := &[]int{}
			in := genOrdInput(out)
			before := runOrdInput(in, out)
			subject(t, in)
			after := runOrdInput(in, out)
			assert.Must(t).Equal(before, after)
		})
	})
}

func TestRandomOrderer_Order(t *testing.T) {
	s := NewSpec(t)
	s.NoSideEffect()

	var (
		seed = Let(s, func(t *T) int64 { return int64(fixtures.Random.Int()) })
		ord  = ord.Let(s, func(t *T) orderer {
			return randomOrderer{Seed: seed.Get(t)}
		})
	)

	subject := func(t *T, in []func()) {
		ord.Get(t).Order(in)
	}

	s.Then(`after the ordering the order of ids list will be shuffled up`, func(t *T) {
		out := &[]int{}
		in := genOrdInput(out)

		before := runOrdInput(in, out)
		subject(t, in) // after ordering
		after := runOrdInput(in, out)

		assert.Must(t).NotEqual(before, after, `after ordering, it should be different`)
	})

	s.Then(`ordering should not affect the length`, func(t *T) {
		out := &[]int{}
		in := genOrdInput(out)
		before := runOrdInput(in, out)
		t.Must.NotEqual(0, len(before))
		subject(t, in) // after ordering
		after := runOrdInput(in, out)
		assert.Must(t).Equal(len(before), len(after))
	})

	s.Then(`the ordering should not affect the content`, func(t *T) {
		out := &[]int{}
		in := genOrdInput(out)
		before := runOrdInput(in, out)
		t.Must.NotEqual(0, len(before))
		subject(t, in) // after ordering
		after := runOrdInput(in, out)
		t.Must.ContainExactly(before, after)
	})

	s.Then(`shuffling should be deterministic and always the same for the same seed`, func(t *T) {
		out := &[]int{}
		ogIn := genOrdInput(out)
		in := cpyOrdInput(ogIn)
		initial := runOrdInput(in, out)

		subject(t, in)
		res1 := runOrdInput(in, out)
		t.Must.ContainExactly(initial, res1)

		in = cpyOrdInput(ogIn) // reset input order
		subject(t, in)         // run again
		res2 := runOrdInput(in, out)
		t.Must.ContainExactly(initial, res2)

		assert.Must(t).Equal(res1, res2, `both outcome of the shuffle should be the same with the same Seed`)
	})

	s.Then(`different seed yield different shuffling`, func(t *T) {
		Retry{Strategy: Waiter{WaitTimeout: time.Second}}.Assert(t, func(it assert.It) {
			out := &[]int{}
			ogIn := genOrdInput(out)
			initial := runOrdInput(ogIn, out)
			seed1 := int64(fixtures.Random.Int())
			seed2 := int64(fixtures.Random.Int())
			it.Must.NotEqual(seed1, seed2, `given the two seed that will be used is different`)

			// random order with a seed
			in := cpyOrdInput(ogIn)

			ord.Set(t, randomOrderer{Seed: seed1})
			subject(t, in)
			res1 := runOrdInput(in, out)
			it.Must.ContainExactly(initial, res1)
			it.Must.NotEqual(initial, res1)

			// random order with different seed
			in = cpyOrdInput(ogIn)
			ord.Set(t, randomOrderer{Seed: seed2})
			subject(t, in)
			res2 := runOrdInput(in, out)
			it.Must.ContainExactly(initial, res2)
			it.Must.NotEqual(initial, res2)

			it.Logf(`the two ordering should be different because the different seeds`)
			// the two random ordering  with different seed because the different seed
			it.Must.NotEqual(res1, res2)
			it.Must.ContainExactly(res1, res2)
		})
	})
}

func TestNewOrderer(t *testing.T) {
	s := NewSpec(t)

	seed := Let(s, func(t *T) int64 {
		return int64(t.Random.Int())
	})
	subject := func(t *T) orderer {
		return newOrderer(t, seed.Get(t))
	}

	s.Before(func(t *T) {
		internal.SetupCacheFlush(t)
	})

	s.When(`mod is unknown`, func(s *Spec) {
		s.Before(func(t *T) {
			SetEnv(t, EnvKeyOrdering, `unknown`)
		})

		s.Then(`it will panic`, func(t *T) {
			t.Must.Panic(func() { subject(t) })
		})
	})

	s.When(`mod is random`, func(s *Spec) {
		s.Before(func(t *T) {
			SetEnv(t, EnvKeyOrdering, string(OrderingAsRandom))
		})

		s.Then(`random orderer provided`, func(t *T) {
			v, ok := subject(t).(randomOrderer)
			t.Must.True(ok)
			t.Must.Equal(seed.Get(t), v.Seed)
		})
	})

	s.When(`mod set ordering as tests are defined`, func(s *Spec) {
		s.Before(func(t *T) {
			SetEnv(t, EnvKeyOrdering, string(OrderingAsDefined))
		})

		s.Then(`null orderer provided`, func(t *T) {
			_, ok := subject(t).(nullOrderer)
			t.Must.True(ok)
		})
	})
}

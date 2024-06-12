package random_test

import (
	"math/rand"
	"testing"

	"go.llib.dev/testcase"
	"go.llib.dev/testcase/let"
	"go.llib.dev/testcase/random"
)

func ExamplePick_randomValuePicking() {
	// Pick randomly from the values of 1,2,3
	var _ = random.Pick(nil, 1, 2, 3)
}

func ExamplePick_pseudoRandomValuePicking() {
	// Pick pseudo randomly from the given values using the seed.
	// This will make picking deterministically random when the same seed is used.
	const seed = 42
	rnd := random.New(rand.NewSource(seed))
	var _ = random.Pick(rnd, "one", "two", "three")
}

func TestPick(t *testing.T) {
	s := testcase.NewSpec(t)

	var (
		rnd = testcase.Let[*random.Random](s, nil)
		vs  = testcase.Let(s, func(t *testcase.T) []int {
			return random.Slice(t.Random.IntB(3, 5), t.Random.Int)
		})
	)
	act := func(t *testcase.T) int {
		return random.Pick[int](rnd.Get(t), vs.Get(t)...)
	}

	thenItWillStillSelectARandomValue := func(s *testcase.Spec) {
		s.Then("it will still select a random value", func(t *testcase.T) {
			var exp = make(map[int]struct{})
			for _, k := range vs.Get(t) {
				exp[k] = struct{}{}
			}

			var got = make(map[int]struct{})
			t.Eventually(func(it *testcase.T) {
				got[act(t)] = struct{}{}

				it.Must.ContainExactly(exp, got)
			})
		})
	}

	s.When("random.Random is nil", func(s *testcase.Spec) {
		rnd.LetValue(s, nil)

		thenItWillStillSelectARandomValue(s)
	})

	s.When("random.Random is supplied", func(s *testcase.Spec) {
		seed := let.IntB(s, 0, 42)
		mkSource := func(t *testcase.T) rand.Source {
			return rand.NewSource(int64(seed.Get(t)))
		}
		rnd.Let(s, func(t *testcase.T) *random.Random {
			return random.New(mkSource(t))
		})

		thenItWillStillSelectARandomValue(s)

		s.Then("random pick is determinstic through controlling the seed", func(t *testcase.T) {
			exp := act(t)
			rnd.Get(t).Source = mkSource(t)
			got := act(t)
			t.Must.Equal(exp, got)
		})
	})
}

package let_test

import (
	"context"
	"testing"
	"time"

	"github.com/adamluzsi/testcase"
	"github.com/adamluzsi/testcase/assert"
	"github.com/adamluzsi/testcase/let"
	"github.com/adamluzsi/testcase/random"
)

func TestSTD_smoke(t *testing.T) {
	s := testcase.NewSpec(t)

	Context := let.Context(s)
	Error := let.Error(s)
	String := let.String(s)
	StringNC := let.StringNC(s, 42, random.CharsetASCII())
	Bool := let.Bool(s)
	Int := let.Int(s)
	IntN := let.IntN(s, 42)
	IntB := let.IntB(s, 7, 42)
	Time := let.Time(s)
	TimeB := let.TimeB(s, time.Now().AddDate(-1, 0, 0), time.Now())
	UUID := let.UUID(s)
	Element := let.ElementFrom[string](s, "foo", "bar", "baz")

	charsterIs := func(t *testcase.T, cs, str string) {
		for _, v := range str {
			t.Must.Contain(cs, string(v))
		}
	}

	s.Test("", func(t *testcase.T) {
		t.Must.Equal(context.Background(), Context.Get(t))
		t.Must.Error(Error.Get(t))
		t.Must.NotEmpty(String.Get(t))
		t.Must.NotEmpty(StringNC.Get(t))
		t.Must.True(42 == len(StringNC.Get(t)))
		charsterIs(t, random.CharsetASCII(), StringNC.Get(t))
		t.Must.NotEmpty(t, Int.Get(t))
		t.Must.NotEmpty(t, IntN.Get(t))
		t.Must.NotEmpty(t, IntB.Get(t))
		t.Must.NotEmpty(t, Time.Get(t))
		t.Must.NotEmpty(t, TimeB.Get(t))
		t.Must.True(TimeB.Get(t).After(time.Now().AddDate(-1, 0, -1)))
		t.Must.NotEmpty(t, UUID.Get(t))
		t.Must.NotEmpty(t, Element.Get(t))
		t.Eventually(func(it assert.It) {
			it.Must.True(Bool.Get(testcase.ToT(&t.TB)))
		})
	})
}

package let_test

import (
	"context"
	"testing"
	"time"

	"go.llib.dev/testcase"
	"go.llib.dev/testcase/assert"
	"go.llib.dev/testcase/let"
	"go.llib.dev/testcase/random"
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
		t.Must.NotNil(Context.Get(t))
		t.Must.NoError(Context.Get(t).Err())
		t.Must.NotWithin(time.Millisecond, func(ctx context.Context) {
			select {
			case <-Context.Get(t).Done():
				// expect to block
			case <-ctx.Done():
				// will be done after the assertion
			}
		})
		t.Must.Error(Error.Get(t))
		t.Must.NotEmpty(String.Get(t))
		t.Must.NotEmpty(StringNC.Get(t))
		t.Must.True(len(StringNC.Get(t)) == 42)
		charsterIs(t, random.CharsetASCII(), StringNC.Get(t))
		t.Must.NotEmpty(Int.Get(t))
		t.Eventually(func(t *testcase.T) {
			t.Must.NotEmpty(IntN.Get(testcase.ToT(&t.TB)))
		})
		t.Must.NotEmpty(IntB.Get(t))
		t.Must.NotEmpty(Time.Get(t))
		t.Must.NotEmpty(TimeB.Get(t))
		t.Must.True(TimeB.Get(t).After(time.Now().AddDate(-1, 0, -1)))
		t.Must.NotEmpty(UUID.Get(t))
		t.Must.NotEmpty(Element.Get(t))
		t.Eventually(func(it *testcase.T) {
			it.Must.True(Bool.Get(testcase.ToT(&t.TB)))
		})
	})
}

func TestContext_cancellationDuringCleanup(t *testing.T) {
	s := testcase.NewSpec(t)
	s.Sequential()
	ctxVar := let.Context(s)
	var ctx context.Context
	s.Test("", func(t *testcase.T) {
		ctx = ctxVar.Get(t)
		t.Must.NoError(ctx.Err())
	})
	s.Finish()
	assert.NotNil(t, ctx)
	assert.ErrorIs(t, context.Canceled, ctx.Err())
}

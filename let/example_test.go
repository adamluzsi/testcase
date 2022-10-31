package let_test

import (
	"github.com/adamluzsi/testcase"
	"github.com/adamluzsi/testcase/let"
	"github.com/adamluzsi/testcase/random"
	"testing"
	"time"
)

func ExampleFirstName() {
	s := testcase.NewSpec((testing.TB)(nil))

	firstName := let.FirstName(s)

	s.Test("", func(t *testcase.T) {
		t.Log(firstName.Get(t))
	})
}

func ExampleLastName() {
	s := testcase.NewSpec((testing.TB)(nil))

	lastName := let.LastName(s)

	s.Test("", func(t *testcase.T) {
		t.Log(lastName.Get(t))
	})
}

func ExampleEmail() {
	s := testcase.NewSpec((testing.TB)(nil))

	email := let.Email(s)

	s.Test("", func(t *testcase.T) {
		t.Log(email.Get(t))
	})
}

func ExampleContext() {
	s := testcase.NewSpec((testing.TB)(nil))

	ctx := let.Context(s)

	s.Test("", func(t *testcase.T) {
		t.Logf("%#v", ctx.Get(t))
	})
}

func ExampleError() {
	s := testcase.NewSpec((testing.TB)(nil))

	expectedErr := let.Error(s)

	s.Test("", func(t *testcase.T) {
		t.Log(expectedErr.Get(t))
	})
}

func ExampleString() {
	s := testcase.NewSpec((testing.TB)(nil))

	str := let.String(s)

	s.Test("", func(t *testcase.T) {
		t.Log(str.Get(t))
	})
}

func ExampleStringNC() {
	s := testcase.NewSpec((testing.TB)(nil))

	str := let.StringNC(s, 42, random.CharsetASCII())

	s.Test("", func(t *testcase.T) {
		t.Log(str.Get(t))
	})
}

func ExampleBool() {
	s := testcase.NewSpec((testing.TB)(nil))

	b := let.Bool(s)

	s.Test("", func(t *testcase.T) {
		t.Log(b.Get(t))
	})
}

func ExampleInt() {
	s := testcase.NewSpec((testing.TB)(nil))

	n := let.Int(s)

	s.Test("", func(t *testcase.T) {
		t.Log(n.Get(t))
	})
}

func ExampleIntN() {
	s := testcase.NewSpec((testing.TB)(nil))

	n := let.IntN(s, 42)

	s.Test("", func(t *testcase.T) {
		t.Log(n.Get(t))
	})
}

func ExampleIntB() {
	s := testcase.NewSpec((testing.TB)(nil))

	n := let.IntB(s, 7, 42)

	s.Test("", func(t *testcase.T) {
		t.Log(n.Get(t))
	})
}

func ExampleTime() {
	s := testcase.NewSpec((testing.TB)(nil))

	tm := let.Time(s)

	s.Test("", func(t *testcase.T) {
		t.Log(tm.Get(t).Format(time.RFC3339))
	})
}

func ExampleTimeB() {
	s := testcase.NewSpec((testing.TB)(nil))

	tm := let.TimeB(s, time.Now().AddDate(-1, 0, 0), time.Now())

	s.Test("", func(t *testcase.T) {
		t.Log(tm.Get(t).Format(time.RFC3339))
	})
}

func ExampleUUID() {
	s := testcase.NewSpec((testing.TB)(nil))

	uuid := let.UUID(s)

	s.Test("", func(t *testcase.T) {
		t.Log(uuid.Get(t))
	})
}

func ExampleElementFrom() {
	s := testcase.NewSpec((testing.TB)(nil))

	v := let.ElementFrom(s, "foo", "bar", "baz")

	s.Test("", func(t *testcase.T) {
		t.Log(v.Get(t))
	})
}

func ExampleWith_func() {
	s := testcase.NewSpec((testing.TB)(nil))

	v := let.With[int](s, func() int {
		return 42
	})

	s.Test("", func(t *testcase.T) {
		t.Log(v.Get(t))
	})
}

func ExampleWith_testingTBFunc() {
	s := testcase.NewSpec((testing.TB)(nil))

	v := let.With[int](s, func(tb testing.TB) int {
		return 42
	})

	s.Test("", func(t *testcase.T) {
		t.Log(v.Get(t))
	})
}

func ExampleWith_testcaseTFunc() {
	s := testcase.NewSpec((testing.TB)(nil))

	v := let.With[int](s, func(t *testcase.T) int {
		return t.Random.Int()
	})

	s.Test("", func(t *testcase.T) {
		t.Log(v.Get(t))
	})
}

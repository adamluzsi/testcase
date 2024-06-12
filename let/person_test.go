package let_test

import (
	"testing"

	"go.llib.dev/testcase"
	"go.llib.dev/testcase/let"
	"go.llib.dev/testcase/random/sextype"
)

func TestPerson_smoke(t *testing.T) {
	s := testcase.NewSpec(t)

	fn := let.FirstName(s)
	ln := let.LastName(s)
	mfn := let.FirstName(s, sextype.Male)
	em := let.Email(s)

	s.Test("", func(t *testcase.T) {
		t.Must.NotEmpty(fn.Get(t))
		t.Must.NotEmpty(ln.Get(t))
		t.Must.NotEmpty(mfn.Get(t))
		t.Must.NotEmpty(em.Get(t))
		t.Eventually(func(it *testcase.T) {
			it.Must.Equal(t.Random.Contact(sextype.Male).FirstName, mfn.Get(t))
		})
	})
}

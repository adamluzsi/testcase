package let_test

import (
	"testing"

	"github.com/adamluzsi/testcase"
	"github.com/adamluzsi/testcase/assert"
	"github.com/adamluzsi/testcase/let"
	"github.com/adamluzsi/testcase/random/sextype"
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
		t.Eventually(func(it assert.It) {
			it.Must.Equal(t.Random.Name().First(sextype.Male), mfn.Get(t))
		})
	})
}

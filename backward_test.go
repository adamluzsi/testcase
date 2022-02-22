package testcase_test

import (
	"testing"

	"github.com/adamluzsi/testcase"
	"github.com/adamluzsi/testcase/fixtures"
)

func TestSpec_Let_andLetValue_backwardCompatibility(t *testing.T) {
	s := testcase.NewSpec(t)

	r1 := fixtures.Random.Int()
	r2 := fixtures.Random.Int()

	v1 := s.Let(`answer`, func(t *testcase.T) interface{} { return r1 })
	v2 := s.LetValue(`count`, r2)

	s.Test(``, func(t *testcase.T) {
		t.Must.Equal(r1, v1.Get(t))
		t.Must.Equal(r2, v2.Get(t))
	})
}

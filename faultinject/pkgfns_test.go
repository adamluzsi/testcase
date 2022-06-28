package faultinject_test

import (
	"context"
	"errors"
	"testing"

	"github.com/adamluzsi/testcase"
	"github.com/adamluzsi/testcase/faultinject"
	"github.com/adamluzsi/testcase/random"
)

func TestCheckFor(t *testing.T) {
	s := testcase.NewSpec(t)

	enabled.Bind(s)

	type FaultTag struct{ ID string }

	var (
		ctxV = testcase.Let(s, func(t *testcase.T) context.Context {
			return context.Background()
		})
		expectedErr = testcase.Let(s, func(t *testcase.T) error {
			return errors.New(t.Random.String())
		})
		targetTag = testcase.Let[faultinject.Tag](s, func(t *testcase.T) faultinject.Tag {
			return FaultTag{ID: t.Random.StringNC(7, random.CharsetAlpha())}
		})
	)
	act := func(t *testcase.T) error {
		return faultinject.CheckFor(ctxV.Get(t), targetTag.Get(t), expectedErr.Get(t))
	}

	WhenFaultInjectIsDisabled := func(s *testcase.Spec) {
		s.When("when fault inject is disabled", func(s *testcase.Spec) {
			enabled.LetValue(s, false)

			s.Then("no error is returned", func(t *testcase.T) {
				t.Must.Nil(act(t))
			})
		})
	}

	s.When("nil context is provided", func(s *testcase.Spec) {
		ctxV.Let(s, func(t *testcase.T) context.Context {
			return nil
		})

		s.Then("it will yield no error", func(t *testcase.T) {
			t.Must.Nil(act(t))
		})
	})

	s.When("no fault is injected", func(s *testcase.Spec) {
		ctxV.Let(s, func(t *testcase.T) context.Context {
			return context.Background()
		})

		s.Then("it will yield no error", func(t *testcase.T) {
			t.Must.Nil(act(t))
		})
	})

	s.When("matching fault Tag is injected", func(s *testcase.Spec) {
		tag := testcase.Let[faultinject.Tag](s, nil)
		ctxV.Let(s, func(t *testcase.T) context.Context {
			return faultinject.Inject(context.Background(), tag.Get(t))
		})

		s.And("tag is matching with the expected target tag", func(s *testcase.Spec) {
			tag.Let(s, func(t *testcase.T) faultinject.Tag {
				return targetTag.Get(t)
			})

			s.Then("error is injected", func(t *testcase.T) {
				t.Must.ErrorIs(expectedErr.Get(t), act(t))
			})

			WhenFaultInjectIsDisabled(s)
		})

		s.And("tag is not matching with the expected target tag", func(s *testcase.Spec) {
			tag.Let(s, func(t *testcase.T) faultinject.Tag {
				return FaultTag{ID: t.Random.StringNC(5, random.CharsetASCII())}
			})

			s.Then("error is not injected", func(t *testcase.T) {
				t.Must.Nil(act(t))
			})
		})
	})

	s.When("generic Fault injected", func(s *testcase.Spec) {
		tagErr := testcase.Let(s, func(t *testcase.T) error {
			return errors.New(t.Random.String())
		})
		ctxV.Let(s, func(t *testcase.T) context.Context {
			return faultinject.Inject(context.Background(), faultinject.Fault{Error: tagErr.Get(t)})
		})

		s.Then("Fault's error is injected", func(t *testcase.T) {
			t.Must.ErrorIs(tagErr.Get(t), act(t))
		})

		WhenFaultInjectIsDisabled(s)
	})
}

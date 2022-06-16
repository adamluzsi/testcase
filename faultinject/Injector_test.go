package faultinject_test

import (
	"context"
	"errors"
	"testing"

	"github.com/adamluzsi/testcase"
	"github.com/adamluzsi/testcase/faultinject"
	"github.com/adamluzsi/testcase/random"
)

func TestInjector(t *testing.T) {
	s := testcase.NewSpec(t)

	injector := testcase.Let(s, func(t *testcase.T) faultinject.Injector {
		return faultinject.Injector{}
	})

	s.Describe(".Check", func(s *testcase.Spec) {
		ctxV := testcase.Let(s, func(t *testcase.T) context.Context { return context.Background() })
		subject := func(t *testcase.T) error {
			return injector.Get(t).Check(ctxV.Get(t))
		}

		s.When("nil context is provided", func(s *testcase.Spec) {
			ctxV.Let(s, func(t *testcase.T) context.Context {
				return nil
			})

			s.Then("it will yield no error", func(t *testcase.T) {
				t.Must.Nil(subject(t))
			})
		})

		s.When("no fault is injected", func(s *testcase.Spec) {
			// nothing to do

			s.Then("it will yield no error", func(t *testcase.T) {
				t.Must.Nil(subject(t))
			})
		})

		s.When("tag is configured with the injector", func(s *testcase.Spec) {
			tagName := testcase.Let(s, func(t *testcase.T) string {
				return t.Random.StringNC(5, random.CharsetAlpha())
			})
			expectedErr := testcase.Let(s, func(t *testcase.T) error {
				return errors.New(t.Random.String())
			})
			s.Before(func(t *testcase.T) {
				injector.Set(t, injector.Get(t).OnTag(tagName.Get(t), expectedErr.Get(t)))
			})

			SpecInjectionCases(s, ctxV, subject, tagName, expectedErr)
		})

		s.When("many tag is configured with the injector", func(s *testcase.Spec) {
			tagName := testcase.Let(s, func(t *testcase.T) string {
				return t.Random.StringNC(5, random.CharsetAlpha())
			})
			expectedErr := testcase.Let(s, func(t *testcase.T) error {
				return errors.New(t.Random.String())
			})
			othTagName := testcase.Let(s, func(t *testcase.T) string {
				return t.Random.StringNC(5, random.CharsetAlpha())
			})
			othExpectedErr := testcase.Let(s, func(t *testcase.T) error {
				return errors.New(t.Random.String())
			})
			s.Before(func(t *testcase.T) {
				injector.Set(t, injector.Get(t).
					OnTag(tagName.Get(t), expectedErr.Get(t)).
					OnTag(othTagName.Get(t), othExpectedErr.Get(t)))
			})

			SpecInjectionCases(s, ctxV, subject, tagName, expectedErr)

			s.And("fault is arranged for the other tag", func(s *testcase.Spec) {
				s.Before(func(t *testcase.T) {
					ctxV.Set(t, faultinject.Inject(ctxV.Get(t), othTagName.Get(t)))
				})

				s.Then("other fault is triggered by the injection", func(t *testcase.T) {
					t.Must.ErrorIs(othExpectedErr.Get(t), subject(t))
				})
			})
		})
	})
}

func SpecInjectionCases(s *testcase.Spec,
	ctxV testcase.Var[context.Context],
	checkSubject func(t *testcase.T) error,
	tagName testcase.Var[string],
	expectedErr testcase.Var[error],
) {
	s.And("fault injected by tag", func(s *testcase.Spec) {
		s.Before(func(t *testcase.T) {
			ctxV.Set(t, faultinject.Inject(ctxV.Get(t), tagName.Get(t)))
		})

		s.Then("it yields the error back on the first call", func(t *testcase.T) {
			t.Must.ErrorIs(expectedErr.Get(t), checkSubject(t))
		})

		s.Then("it yields no error after the fault is already retrieved", func(t *testcase.T) {
			_ = checkSubject(t)
			for i, probeCount := 0, t.Random.IntB(3, 7); i < probeCount; i++ {
				t.Must.Nil(checkSubject(t))
			}
		})

		s.And("multiple fault is arranged for the same tag", func(s *testcase.Spec) {
			s.Before(func(t *testcase.T) {
				ctxV.Set(t, faultinject.Inject(ctxV.Get(t), tagName.Get(t)))
			})

			s.Then("it yields back an error on the first call", func(t *testcase.T) {
				t.Must.ErrorIs(expectedErr.Get(t), checkSubject(t))
			})

			s.Then("on sequential execution, it returns back the faults in a FIFO order until all fault are consumed", func(t *testcase.T) {
				t.Must.ErrorIs(expectedErr.Get(t), checkSubject(t))
				t.Must.ErrorIs(expectedErr.Get(t), checkSubject(t))
				t.Must.Nil(checkSubject(t))
			})
		})
	})

	s.And("the tag name does not matches", func(s *testcase.Spec) {
		othTagName := testcase.Let(s, func(t *testcase.T) string {
			return t.Random.StringNC(5, random.CharsetAlpha())
		})
		s.Before(func(t *testcase.T) {
			ctxV.Set(t, faultinject.Inject(ctxV.Get(t), othTagName.Get(t)))
		})

		s.Then("it yields no error", func(t *testcase.T) {
			t.Must.Nil(checkSubject(t))
		})
	})
}
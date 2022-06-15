package faultinject_test

import (
	"context"
	"errors"
	"testing"

	"github.com/adamluzsi/testcase"
	"github.com/adamluzsi/testcase/faultinject"
)

func TestContext(t *testing.T) {
	s := testcase.NewSpec(t)
	defer s.Finish()

	s.Describe("#Check", func(s *testcase.Spec) {
		ctxV := testcase.Let(s, func(t *testcase.T) context.Context { return context.Background() })
		tags := testcase.Let(s, func(t *testcase.T) []string { return nil })
		subject := func(t *testcase.T) error {
			return faultinject.Check(ctxV.Get(t), tags.Get(t)...)
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

		s.When("fault injected by fault-tag", func(s *testcase.Spec) {
			faultV1 := testcase.Let(s, func(t *testcase.T) faultinject.Fault {
				return faultinject.Fault{
					OnTag: t.Random.String(),
					Error: errors.New(t.Random.String()),
				}
			})

			s.Before(func(t *testcase.T) {
				ctxV.Set(t, faultinject.Inject(ctxV.Get(t), faultV1.Get(t)))
			})

			s.And("the fault tag is supplied to the Check function", func(s *testcase.Spec) {
				tags.Let(s, func(t *testcase.T) []string {
					var ns []string
					for i, max := 0, t.Random.IntB(5, 7); i < max; i++ {
						ns = append(ns, t.Random.String())
					}
					ns = append(ns, faultV1.Get(t).OnTag)
					return ns
				})

				s.Then("it yields the error back on the first call", func(t *testcase.T) {
					t.Must.ErrorIs(faultV1.Get(t).Error, subject(t))
				})

				s.Then("it yields no error after the fault is already retrieved", func(t *testcase.T) {
					_ = subject(t)
					for i, probeCount := 0, t.Random.IntB(3, 7); i < probeCount; i++ {
						t.Must.Nil(subject(t))
					}
				})

				s.And("multiple fault is arranged for the .Check call", func(s *testcase.Spec) {
					faultV2 := testcase.Let(s, func(t *testcase.T) faultinject.Fault {
						return faultinject.Fault{
							OnTag: faultV1.Get(t).OnTag,
							Error: errors.New(t.Random.String()),
						}
					})

					s.Before(func(t *testcase.T) {
						ctxV.Set(t, faultinject.Inject(ctxV.Get(t), faultV2.Get(t)))
					})

					s.Then("it yields back the first error on the first call", func(t *testcase.T) {
						t.Must.ErrorIs(faultV1.Get(t).Error, subject(t))
					})

					s.Then("on sequential execution, it returns back the faults in a FIFO order until all fault are consumed", func(t *testcase.T) {
						t.Must.ErrorIs(faultV1.Get(t).Error, subject(t))
						t.Must.ErrorIs(faultV2.Get(t).Error, subject(t))
						t.Must.Nil(subject(t))
					})
				})
			})

			s.And("the fault-tag is not passed to the tags", func(s *testcase.Spec) {
				tags.Let(s, func(t *testcase.T) []string {
					var ns []string
					for i, max := 0, t.Random.IntB(5, 7); i < max; i++ {
						ns = append(ns, t.Random.String())
					}
					return ns
				})

				s.Then("it yields the error back error", func(t *testcase.T) {
					t.Must.Nil(subject(t))
				})
			})
		})

		s.When("fault injected by function name", func(s *testcase.Spec) {
			funcName := testcase.Let[string](s, nil)
			faultV1 := testcase.Let(s, func(t *testcase.T) faultinject.Fault {
				return faultinject.Fault{
					OnFunc: funcName.Get(t),
					Error:  errors.New(t.Random.String()),
				}
			})

			s.Before(func(t *testcase.T) {
				ctxV.Set(t, faultinject.Inject(ctxV.Get(t), faultV1.Get(t)))
			})

			s.And("the function name matches with the caller", func(s *testcase.Spec) {
				funcName.LetValue(s, "faultinject_test.TestContext")

				s.Then("it yields the error back on the first call", func(t *testcase.T) {
					t.Must.ErrorIs(faultV1.Get(t).Error, subject(t))
				})

				s.Then("it yields no error after the fault is already retrieved", func(t *testcase.T) {
					_ = subject(t)
					for i, probeCount := 0, t.Random.IntB(3, 7); i < probeCount; i++ {
						t.Must.Nil(subject(t))
					}
				})

				s.And("multiple fault is arranged for the .Check call", func(s *testcase.Spec) {
					faultV2 := testcase.Let(s, func(t *testcase.T) faultinject.Fault {
						return faultinject.Fault{
							OnFunc: funcName.Get(t),
							Error:  errors.New(t.Random.String()),
						}
					})

					s.Before(func(t *testcase.T) {
						ctxV.Set(t, faultinject.Inject(ctxV.Get(t), faultV2.Get(t)))
					})

					s.Then("it yields back the first error on the first call", func(t *testcase.T) {
						t.Must.ErrorIs(faultV1.Get(t).Error, subject(t))
					})

					s.Then("on sequential execution, it returns back the faults in a FIFO order until all fault are consumed", func(t *testcase.T) {
						t.Must.ErrorIs(faultV1.Get(t).Error, subject(t))
						t.Must.ErrorIs(faultV2.Get(t).Error, subject(t))
						t.Must.Nil(subject(t))
					})
				})
			})

			s.And("the function name does not matches with the caller", func(s *testcase.Spec) {
				funcName.LetValue(s, "faultinject_test.NotTestContext")

				s.Then("it yields the error back error", func(t *testcase.T) {
					t.Must.Nil(subject(t))
				})
			})
		})
	})
}

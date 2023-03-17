package testcase

import (
	"fmt"
	"sort"
)

// TableTest allows you to make table tests, without the need to use a boilerplate.
// It optionally allows to use a Spec instead of a testing.TB,
// and then the table tests will inherit the Spec context.
// It guards against mistakes such as using for+t.Run+t.Parallel without variable shadowing.
// TableTest allows a variety of use, please check examples for further information on that.
func TableTest[TBS anyTBOrSpec, TC sBlock | tBlock | any, Act tBlock | sBlock | func(*T, TC)](
	tbs TBS,
	tcs map[ /* description */ string]TC,
	act Act,
) {
	s := ToSpec(tbs)
	var tests []tableTestTestCase[TC]
	for desc, tc := range tcs {
		tests = append(tests, tableTestTestCase[TC]{
			Desc: desc,
			TC:   tc,
		})
	}
	sort.Slice(tests, func(i, j int) bool {
		return tests[i].Desc < tests[j].Desc
	})
	runT := func(s *Spec, test tableTestTestCase[TC], act func(t *T, tc TC)) {
		switch tc := any(test.TC).(type) {
		case sBlock:
			s.Context(test.Desc, func(s *Spec) {
				tc(s)
				s.Test("", func(t *T) {
					act(t, test.TC)
				})
			})
		case tBlock:
			s.Context(test.Desc, func(s *Spec) {
				s.Before(tc)
				s.Test("", func(t *T) {
					act(t, test.TC)
				})
			})
		default:
			s.Test(test.Desc, func(t *T) {
				act(t, test.TC)
			})
		}
	}
	runS := func(s *Spec, test tableTestTestCase[TC], act sBlock) {
		switch tc := any(test.TC).(type) {
		case sBlock:
			s.Context(test.Desc, func(s *Spec) {
				tc(s)
				act(s)
			})
		case tBlock:
			s.Context(test.Desc, func(s *Spec) {
				s.Before(tc)
				act(s)
			})
		default:
			panic(fmt.Sprintf("unsuported TableTest setup: TC<%T> <-> Act<%T>", test.TC, act))
		}
	}
	s.Context("", func(s *Spec) {
		for _, test := range tests {
			test := test // pass by value copy to avoid funny concurrency issues
			switch act := any(act).(type) {
			case sBlock:
				runS(s, test, act)
			case tBlock:
				runT(s, test, func(t *T, tc TC) { act(t) })
			case func(*T, TC):
				runT(s, test, act)
			}
		}
	})
}

type tableTestTestCase[TC any] struct {
	Desc string
	TC   TC
}

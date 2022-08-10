package testcase

import "sort"

// TableTest allows you to make table tests, without the need to use a boilerplate.
// It optionally allows to use a Spec instead of a testing.TB,
// and then the table tests will inherit the Spec context.
// It also ensures that the
func TableTest[TC sBlock | tBlock | any, Act tBlock | func(*T, TC)](
	tbOrSpec any,
	tcs map[ /* description */ string]TC,
	act Act,
) {
	s := toSpec(tbOrSpec)
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
	actFn := func(t *T, tc TC) {
		switch fn := any(act).(type) {
		case tBlock:
			fn(t)
		case func(*T, TC):
			fn(t, tc)
		}
	}
	for _, test := range tests {
		test := test // pass by value copy to avoid funny concurrency issues
		switch tc := any(test.TC).(type) {
		case sBlock:
			s.Context(test.Desc, func(s *Spec) {
				tc(s)
				s.Test("", func(t *T) {
					actFn(t, test.TC)
				})
			})

		case tBlock:
			s.Context(test.Desc, func(s *Spec) {
				s.Before(tc)
				s.Test("", func(t *T) {
					actFn(t, test.TC)
				})
			})

		default:
			s.Test(test.Desc, func(t *T) {
				actFn(t, test.TC)
			})
		}
	}
}

type tableTestTestCase[TC any] struct {
	Desc string
	TC   TC
}

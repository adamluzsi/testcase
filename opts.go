package testcase

import (
	"fmt"
	"time"
)

// Flaky will mark the spec/test as unstable.
// Flaky test execution is tolerant towards failing assertion
// and these tests will be rerun in case of a failure.
// A Wait Timeout for a successful flaky test must be provided.
//
// The primary use-case is that when a team focus on shipping out the value,
// and time is short till deadlines.
// These flaky tests prevent CI/CD pipelines often turned off in the heat of the moment to let pass the latest changes.
// The motivation behind is to gain time for the team to revisit these tests after the release and then learn from it.
// At the same time, they intend to fix it as well.
// These tests, however often forgotten, and while they are not the greatest assets of the CI pipeline,
// they often still serve essential value.
//
// As a Least wrong solution, instead of skipping these tests, you can mark them as flaky, so in a later time,
// finding these flaky tests in the project should be easy.
// When you flag a test as flaky, you must provide a timeout value that will define a testing time window
// where the test can be rerun multiple times by the framework.
// If the test can't run successfully within this time-window, the test will fail.
// This failure potentially means that the underlying functionality is broken,
// and the committer should reevaluate the changes in the last commit.
//
// While this functionality might help in tough times,
// it is advised to pair the usage with a scheduled monthly CI pipeline job.
// The Job should check the testing code base for the flaky flag.
//
func Flaky(CountOrTimeout interface{}) SpecOption {
	switch n := CountOrTimeout.(type) {
	case time.Duration:
		return Retry{Strategy: Waiter{WaitTimeout: n}}
	case int:
		return Retry{Strategy: RetryCount(n)}
	default:
		panic(fmt.Errorf(`%T is not supported by Flaky flag`, CountOrTimeout))
	}
}

//func Timeout(duration time.Duration) SpecOption {}
//func OrderWith(orderer) SpecOption {}

func SkipBenchmark() SpecOption {
	return specOptionFunc(func(c *Spec) {
		c.skipBenchmark = true
	})
}

// Group creates a testing group in the specification.
// During test execution, a group will be bundled together,
// and parallel tests will run concurrently within the the testing group.
func Group(name string) SpecOption {
	return specOptionFunc(func(s *Spec) {
		s.group = name
	})
}

func parallel() SpecOption {
	return specOptionFunc(func(s *Spec) {
		s.parallel = true
	})
}

func sequential() SpecOption {
	return specOptionFunc(func(s *Spec) {
		s.sequential = true
	})
}

////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

type SpecOption interface {
	setup(*Spec)
}

type specOptionFunc func(*Spec)

func (fn specOptionFunc) setup(s *Spec) { fn(s) }

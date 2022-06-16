package testcase

import (
	"fmt"
	"time"

	"github.com/adamluzsi/testcase/assert"
)

// Flaky will mark the spec/testCase as unstable.
// Flaky testCase execution is tolerant towards failing assertion
// and these tests will be rerun in case of a failure.
// A Wait Timeout for a successful flaky testCase must be provided.
//
// The primary use-case is that when a team focus on shipping orderingOutput the value,
// and time is short till deadlines.
// These flaky tests prevent CI/CD pipelines often turned off in the heat of the moment to let pass the latest changes.
// The motivation behind is to gain time for the team to revisit these tests after the release and then learn from it.
// At the same time, they intend to fix it as well.
// These tests, however often forgotten, and while they are not the greatest assets of the CI pipeline,
// they often still serve essential value.
//
// As a Least wrong solution, instead of skipping these tests, you can mark them as flaky, so in a later time,
// finding these flaky tests in the project should be easy.
// When you flag a testCase as flaky, you must provide a timeout value that will define a testing time window
// where the testCase can be rerun multiple times by the framework.
// If the testCase can't run successfully within this time-window, the testCase will fail.
// This failure potentially means that the underlying functionality is broken,
// and the committer should reevaluate the changes in the last commit.
//
// While this functionality might help in tough times,
// it is advised to pair the usage with a scheduled monthly CI pipeline job.
// The Job should check the testing code base for the flaky flag.
//
func Flaky(CountOrTimeout interface{}) SpecOption {
	retry, ok := makeEventually(CountOrTimeout)
	if !ok {
		panic(fmt.Errorf(`%T is not supported by Flaky flag`, CountOrTimeout))
	}
	return specOptionFunc(func(s *Spec) {
		s.flaky = &retry
	})
}

func makeEventually(i any) (assert.Eventually, bool) {
	switch n := i.(type) {
	case time.Duration:
		return assert.Eventually{RetryStrategy: assert.Waiter{Timeout: n}}, true
	case int:
		return assert.Eventually{RetryStrategy: assert.RetryCount(n)}, true
	case assert.RetryStrategy:
		return assert.Eventually{RetryStrategy: n}, true
	case assert.Eventually:
		return n, true
	default:
		return assert.Eventually{}, false
	}
}

func RetryStrategyForEventually(strategy assert.RetryStrategy) SpecOption {
	return specOptionFunc(func(s *Spec) {
		s.eventually = &assert.Eventually{RetryStrategy: strategy}
	})
}

//func Timeout(duration time.Duration) SpecOption {}
//func OrderWith(orderer) SpecOption {}

func SkipBenchmark() SpecOption {
	return specOptionFunc(func(c *Spec) {
		c.skipBenchmark = true
	})
}

// Group creates a testing group in the specification.
// During testCase execution, a group will be bundled together,
// and parallel tests will run concurrently within the the testing group.
func Group(name string) SpecOption {
	return specOptionFunc(func(s *Spec) {
		s.group = &struct{ name string }{name: name}
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

type SpecOption interface {
	setup(*Spec)
}

type specOptionFunc func(s *Spec)

func (fn specOptionFunc) setup(s *Spec) { fn(s) }

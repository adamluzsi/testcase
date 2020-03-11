<!-- START doctoc generated TOC please keep comment here to allow auto update -->
<!-- DON'T EDIT THIS SECTION, INSTEAD RE-RUN doctoc TO UPDATE -->
**Table of Contents**

- [testcase](#testcase)
  - [Example](#example)
  - [The Case Study about what problem the package solve](#the-case-study-about-what-problem-the-package-solve)
    - [The Problem](#the-problem)
    - [The Requirements](#the-requirements)
    - [The Starting Point](#the-starting-point)
    - [The Initial Implementation](#the-initial-implementation)
    - [A/B testing in the package vision](#ab-testing-in-the-package-vision)
    - [The Current Implementation](#the-current-implementation)
  - [Stability](#stability)
  - [Documentations](#documentations)
  - [Reference Project](#reference-project)

<!-- END doctoc generated TOC please keep comment here to allow auto update -->

[![Mentioned in Awesome Go](https://awesome.re/mentioned-badge.svg)](https://github.com/avelino/awesome-go)
[![GoDoc](https://godoc.org/github.com/adamluzsi/testcase?status.png)](https://godoc.org/github.com/adamluzsi/testcase)
[![Build Status](https://travis-ci.org/adamluzsi/testcase.svg?branch=master)](https://travis-ci.org/adamluzsi/testcase)
[![Go Report Card](https://goreportcard.com/badge/github.com/adamluzsi/testcase)](https://goreportcard.com/report/github.com/adamluzsi/testcase)
[![codecov](https://codecov.io/gh/adamluzsi/testcase/branch/master/graph/badge.svg)](https://codecov.io/gh/adamluzsi/testcase)
# testcase

The `testcase` package provides tooling to apply BDD testing conventions.

## Example

[The examples managed in godoc, please read the documentation example section for more.](https://godoc.org/github.com/adamluzsi/testcase#pkg-examples)

A Basic example:

```go
package mypkg_test

import (
	"testing"

	"github.com/you/mypkg"

	"github.com/stretchr/testify/require"

	"github.com/adamluzsi/testcase"
)

func TestMyType(t *testing.T) {
	s := testcase.NewSpec(t)

	s.NoSideEffect()

	myType := func(t *testcase.T) *mypkg.MyType {
		return &mypkg.MyType{}
	}

	s.Describe(`#IsLower`, func(s *testcase.Spec) {
		var subject = func(t *testcase.T) bool {
			return myType(t).IsLower(t.I(`input`).(string))
		}

		s.When(`input has upcase letter`, func(s *testcase.Spec) {
			s.LetValue(`input`, `UPPER`)

			s.Then(`it will be false`, func(t *testcase.T) {
				require.False(t, subject(t))
			})
		})

		s.When(`input is all lowercase letter`, func(s *testcase.Spec) {
			s.LetValue(`input`, `lower`)

			s.Then(`it will be true`, func(t *testcase.T) {
				require.True(t, subject(t))
			})
		})
	})
}
```

## The Case Study about what problem the package solve

### The Problem

The software architecture used in the projects where I work has decoupled responsibility.
Therefore the layer that deals with managing business entities in an external resource,
a layer that composes these interactors/consumer that manage business entities to form business rules and business logic,
a layer that supply external resource interactions
and a layer that interacts with an external interface such as HTTP later.

Because of this segregation, without a proper BDD testing approach,
there would be a lot of mock tests that would reduce the maintainability of the project.
To conquer this problem, the tests try to use real components instead of a mock, and each higher-level layer (by architecture meaning) that interacts with a layer below, they interact with an interactor that manages a resource to set up a test environment.
This approach introduced a lot of boilerplate, as sharing such context setup added a lot of repetition in tests that use the same components.

To make this less abstract, imagine the following directory layout:
```
.
├── extintf
│   ├── caches               // cache external resource suppliers
│   ├── httpintf             // HTTP external interface implementation
│   ├── queues               // queue external resource suppliers
│   └── storages             // storage external resource supplier
├── services
│   ├── domain-service-1
│   ├── domain-service-2
│   └── domain-service-3
└── usecases
    ├── SignIn.go            // example use-case for sign in
    ├── SignUp.go            // example use-case for sign up
    └── SomeOtherUseCase.go  // example use case
```

And under the services directory, each directory represents certain domain rules and entities.
So when tests needed to be made on a higher level than the domain rules, like the use cases, 
it was crucial to avoid any implementation level detail check in the assertions to test the high-level business rules.
 
So for example, if a domain service particular entity use storage to persist and retrieve domain entities,
it was essential to avoid checking domain entities in the storage used by the domain rules,
when we create tests for a high level in the use-cases interactor.

On lower level for creating the domain rule tests, it was a concern to create tests,
that reflect the expected system behavior instead of the implementation of a specific rule, through mocking dependencies.

### The Requirements

The following requirements were specified for the project to address the issues mentioned above.

- DRY specifications for similar edge cases that enhance the maintainability aspect of the tests/specs.
- shareable helper functions that can improve the readability of high-level tests.
- make the feedback loop as fast as possible to allow small quick changes in the codebase
  * for example, when a test has no side effect to the program, it can be run in parallel
  * so if you don't use global variables (`os Environment variables`) in your currently tested code's scope, then all your tests should run on a separate core. 
- define test subjects with inputs that is not yet provided
  * This allows creating specification where each input needs to be specified explicitly,
    and not defined inputs can be easily seen while making the specification.
  * This allows us to setup steps like "something is done with something else by n", and then later define this value at a test context.
- running a specification should generate a humanly readable specification 
  that helps to build the mental model of a given code in the subject.
- low maintainability cost on the framework side
  * stable API
  * no breaking change 
- test edge cases can be executed alone as well
  * Should be used together with [dlv](https://github.com/go-delve/delve)  
- can visualize code complexity by the spec coverage size

### The Starting Point

So as a starting point, various BDD testing framework projects were checked out to see if there would be a good match.
There were various frameworks already, and each covered some of the critical requirements to solve the issues at hand,
but none answered them all.

Then battle-tested testing frameworks from other languages were checked for inspiration basis.
The high-level approach of the [rspec](https://github.com/rspec/rspec) framework turned out to cover most of the requirements, 
so all that was needed is to extend these core functionality to solve the remaining requirements.

### The Initial Implementation

Initially, two implementations were made.

The `Spec` approach meant to push the test writer to define each test variable and each edge case context with documentation.
This was achieved by providing a structure that helps applying BDD conventions, through a minimal set of helper functions.

The Other approach was the `Steps` which was basically to build a list of function that represented testing steps,
and meant to be used with nested tests through the usage of `testing.T#Run` function.
This approach allowed to use `testing` package nesting strategy mainly,
while composing testing hooks to express a certain `testing.T#Run` scope test runtime context.  

### A/B testing in the package vision

These two approaches then was A/B tested in different projects.
The A/B testing was ongoing for slightly more than 10 months.

In the end of the A/B testing, the following observations were made:
- `Spec` based project tests were more maintainable from the `Steps` based approach.
- `Steps` required less initial effort to learn it.
- `Steps` often encountered cases where variable setup was not possible in isolation.
- projects with `Spec` had generally faster feedback loops.
- tests with `Spec` had better readability when values with teardown were needed.
- `Spec` had advantage in solving testing feature needs commonly without need to change project specifications.
- Using shared specification helpers with `Steps` was messy and hard to refactor.
- `Spec` allowed the same or really similar conventions that community built for `rspec`, `jasmin`, `rtl` and similar testing frameworks.

At the end of the A/B testing, the `Spec` approach turned out to be more preferred in the projects in subject.
You can see a usage of `Spec` approach in an open source project that is called [toggler](https://github.com/toggler-io/toggler).
The specification were initially made there with the MVP set of the `Spec` approach,
therefore not all the latest idiom were applied.

### The Current Implementation

Currently, the `Spec` approach is maintained in the `test case` package,
minor changes tagged, and in generally, rolling release strategy is used.

The internals of the `Spec` is based on the `testing.T#Run` function,
and, as such, the essential parts maintained by the core `testing` package, since the `Spec` package only wraps it. Tests are made to ensure the stability of the implementation in case the `testing` package used parts behavior would change. 
 
## Stability

The package considered stable.
No breaking change is expected to the package exported API.
The package use rolling release conventions to introduce new functionality.

## Documentations

The package detailed documentation is kept in the [GoDoc](https://godoc.org/github.com/adamluzsi/testcase).
* [Nesting guide](/docs/nesting.md)
* [additional docs](/docs)

## Reference Project

* [toggler project, scalable feature toggles on budget for startups](https://github.com/adamluzsi/toggler)
* [frameless project, for a vendor lock free software architecture](https://github.com/adamluzsi/frameless)


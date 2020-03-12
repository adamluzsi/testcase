<!-- START doctoc generated TOC please keep comment here to allow auto update -->
<!-- DON'T EDIT THIS SECTION, INSTEAD RE-RUN doctoc TO UPDATE -->
**Table of Contents**

- [testcase](#testcase)
  - [Documentation](#documentation)
  - [Example](#example)
  - [Summary](#summary)
    - [DRY](#dry)
    - [Modularization](#modularization)
  - [Stability](#stability)
  - [Case Study Of The Package Origin](#case-study-of-the-package-origin)
    - [The Problem](#the-problem)
    - [The Requirements](#the-requirements)
    - [The Starting Point](#the-starting-point)
    - [The Initial Implementation](#the-initial-implementation)
    - [A/B Testing For The Package Vision](#ab-testing-for-the-package-vision)
    - [The Current Implementation](#the-current-implementation)
  - [Reference Project](#reference-project)

<!-- END doctoc generated TOC please keep comment here to allow auto update -->

[![Mentioned in Awesome Go](https://awesome.re/mentioned-badge.svg)](https://github.com/avelino/awesome-go)
[![GoDoc](https://godoc.org/github.com/adamluzsi/testcase?status.png)](https://godoc.org/github.com/adamluzsi/testcase)
[![Build Status](https://travis-ci.org/adamluzsi/testcase.svg?branch=master)](https://travis-ci.org/adamluzsi/testcase)
[![Go Report Card](https://goreportcard.com/badge/github.com/adamluzsi/testcase)](https://goreportcard.com/report/github.com/adamluzsi/testcase)
[![codecov](https://codecov.io/gh/adamluzsi/testcase/branch/master/graph/badge.svg)](https://codecov.io/gh/adamluzsi/testcase)
# testcase

The `testcase` package provides tooling to apply BDD testing conventions.

## [Documentation](https://godoc.org/github.com/adamluzsi/testcase)

[The Official Package documentation managed in godoc](https://godoc.org/github.com/adamluzsi/testcase).

This `README.md` serves as a high level intro into the package, 
and a case study why the package was made.
For package API, examples and usage details about the `testcase` package, 
please see the package [godoc](https://godoc.org/github.com/adamluzsi/testcase).

[additional documentations](/docs):
* [Nesting guide](/docs/nesting.md)

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

## Summary

### DRY

`testcase` provides a way to express common Arrange, Act sections for the Asserts with DRY principle in mind.

- First you can define your Act section with a method under test as the subject of your test specification
    * The Act section invokes the method under test with the arranged parameters.
- Then you can build the context of the Act by Arranging the inputs later with humanly explained reasons
    * The Arrange section initializes objects and sets the value of the data that is passed to the method under test.   
- And lastly you can define the test expected outcome in an Assert section.
    * The Assert section verifies that the action of the method under test behaves as expected. 

Then adding an additional test edge case to the testing suite becomes easier,
as it will have a concrete place where it must be placed.

And if during the creation of the specification, an edge case turns out to be YAGNI,
it can be noted, so visually it will be easier to see what edge case is not specified for the given subject.

The value it gives is that to build test for a certain edge case, 
the required mental model size to express the context becomes smaller,
as you only have to focus on one Arrange at a time,
until you fully build the bigger picture.

It also implicitly visualize the required mental model of your production code by the nesting.
[You can read more on that in the nesting section](/docs/nesting.md).  

### Modularization

On top of the DRY convention, any time you need to Arrange a common scenario about your projects domain event,
you can modularize these setup blocks in a helper functions.

This helps the readability of the test, while keeping the need of mocks to the minimum as possible for a given test.
As a side effect, integration tests can become low hanging fruit for the project.

e.g.:
```go
package mypkg_test

import (
	"testing"

	"my/project/mypkg"


	"github.com/adamluzsi/testcase"

	. "my/project/testing/pkg"
)

func TestMyTypeMyFunc(t *testing.T) {
	s := testcase.NewSpec(t)

	// high level Arrange helpers from my/project/testing/pkg
	SetupSpec(s)
	GivenWeHaveUser(s, `myuser`)
	// .. other givens

	myType := func() *mypkg.MyType { return &mypkg.MyType{} }

	s.Describe(`#MyFunc`, func(s *testcase.Spec) {
		var subject = func(t *testcase.T) { myType().MyFunc(t.I(`myuser`).(*mypkg.User)) } // Act

		s.Then(`edge case description`, func(t *testcase.T) {
			// Assert
			subject(t)
		})
	})
}
```

## Stability

- The package considered stable.
- The package use rolling release conventions.
- No breaking change is planned to the package exported API.
- The package used for production development.
- The package API is only extended if the practical use case proves its necessity.

## Case Study Of The Package Origin

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


  
* The design of the testing lib should not weight more the value of fancy DSL, than golang idioms.
* allow me to run test cases in concurrent execution for specification where I know that no side effect expected.
  * this is especially important me, because I love quick test feedback loops
* allow me to define variables in a way that
    * they receive concrete value later
    * they can be safely overwritten with nested scopes
* strictly regulated usage,
    * with early errors/panics about potential misuse
* I want to use [stretchr/testify](https://github.com/stretchr/testify), so assertions not necessary for me
  * or more precisely, I needed something that guaranteed to allow me the usage of that pkg


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
- specific edge cases can be executed alone easily
  * can be used easily with [dlv](https://github.com/go-delve/delve)
  * can be used with `go test` command out of the box
    * includes `-run` option to specify test case(s)
- can visualize code complexity by the testing specification
- don't build and use testing package level globals

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

### A/B Testing For The Package Vision

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

The `Spec` approach was kept and will be maintained in the `testcase` package.

The internals of the `Spec` is based on the `testing.T#Run` function,
and, as such, the essential parts maintained by the core `testing` package, since the `Spec` package only wraps it.

Tests coverage made to ensure the behavior of the `Spec` approach implementation.
The coverage is more about the behavior, than the code execution flow,
and while some implementation may overlap by tests,
the behavior is defined for each edge case explicitly.

## Reference Project

- [toggler project, scalable feature toggles on budget for startups](https://github.com/adamluzsi/toggler)
- [frameless project, for a vendor lock free software architecture](https://github.com/adamluzsi/frameless)
- [gorest, a minimalist REST controller for go projects](https://github.com/adamluzsi/gorest)


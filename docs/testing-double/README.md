<!-- START doctoc generated TOC please keep comment here to allow auto update -->
<!-- DON'T EDIT THIS SECTION, INSTEAD RE-RUN doctoc TO UPDATE -->

- [Testing Doubles](#testing-doubles)
  - [Dummy](#dummy)
  - [Fake](#fake)
  - [Stub](#stub)
  - [Spy](#spy)
  - [Mock](#mock)

<!-- END doctoc generated TOC please keep comment here to allow auto update -->

# Testing Doubles

## Dummy

Dummy objects are values passed around but never actually used.
They meant to fill parameter lists.

**PRO**

- they can be anything

**CON**

- TBD

**Use**

- fill parameter lists

**Example**

```go
mypkg.Function("dummy", "value")
``` 

## Fake

A Fake is a working implementation,
but usually take some shortcut which makes them not suitable for production.

Fakes are suppliers with working implementations but not the same as the production ones.
Usually, they take some shortcuts and have simplified versions of production code.
The proper fake implementation is also compliant with the [contract](/docs/contracts.md) a role interface has, like the production one.

An example of this shortcut can be an in-memory implementation of a Repository role interface.
This fake implementation will not engage an actual database
but will use a simple collection to store data.

This approach allows us to do integration-testing of services without starting up a database and performing time-consuming requests.

Apart from testing, fake implementation can come in handy for prototyping and spikes.
For example, we can quickly implement and run our system with an in-memory store,
deferring decisions about what technology and concrete design should be used.

Fakes can simplify local development when working with complex external systems.

Example use cases:
- a payment system that always returns with successful payment and does the callback automatically on request.
- email verification process calls verify callback instead of sending an email out.
- [in-memory database for testing](https://martinfowler.com/bliki/InMemoryTestDatabase.html)

**PRO**

- you can start developing your business rules without the need to choose a technology stack ahead of time before you know your business requirements.
- can support easier local development in manual testing and integration tests.
  Fourth, - allowing testing suite optimizations when using real components drastically increases the testing feedback loop time.
  Finally, - it allows taking shortcuts instead of using the concrete external resources when the application runs locally for development purposes.

**CON**

- using fake without an [role interface contract](/docs/contracts.md) introduce manual maintenance costs.
- neglecting to keep fake in sync with the production variant will risk violating dev/prod parity in the project's testing suite.

**Use**

-  test happy-path with it
- replace real implementation in tests when the feedback loop with it is too slow
- test business logic with it

**Example**

[Example fake implementation](/docs/testing-double/fake_test.go) for the [example role interface + contract](/docs/testing-double/spec_helper_test.go).

## Stub

Stub provides canned answers to calls made during the test,
usually not responding to anything outside what's programmed in for the test.

Method Stubbing within the stub allows you to manipulate one or two methods to inject mostly errors to test rainy paths with it.

My suggestion is only to stub a method to fault inject,
and avoid representing a happy path with it whenever possible.

**PRO**

- relatively easy to use
- ideal to inject error with it

**CON**

- when stub testing double used for representing a happy path, we need to introduce a manual chore activity
  to the project to ensure the stub content is up to date with the production

**Use**

- fault injection through monkey patching with embedding

**Example**

- [Stub Object](/docs/testing-double/stub_test.go)
- [Stub One method on a real Object for fault injection](/docs/testing-double/stub_method_test.go)

## Spy
Spy stubs also record information based on how the code called their methods.

Often used to verify "indirect output" of the tested code
by asserting the expectations afterwards,
without having defined the expectations before the tested code is executed.
In addition, it helps in recording information about the indirect object created.

**PRO**

- everything true to stub
- can help to debug

**CON**

- everything that is true to stub
- a risk that the test will focus on implementation details if misused.

**Use**

- checking retry logic behavioural requirements from an analytical point of view

## Mock

Mock is pre-programmed with expectations, forming a specification of the calls they are expected to receive.
They can throw an exception if they receive a call they don't expect
and are checked during verification to ensure they got all the calls they were expecting.

Mocks shine the most when used for large teams where different parts of the code develop in parallel.
After an initial agreement between members about the interface and high-level behaviour between components,
they can start to develop without anything concrete.
This approach cost architecture flexibility by introducing tech debt in the testing suite,
but this debt can be fixed later by cleaning up tests
when all the components are already integrated into the system.

Another example of using mocks is when the project has too much entropy,
and testing with real components would require way too much extra effort.
The extra effort would not stop with only fixing the code
but most likely involve additional rounds of knowledge sharing in the team.
When this is impossible, it is "acceptable" to test implementation details with mocks
while ensuring the behaviour requirements of the role interface being mocked are kept at a minimum.
Ideally, try to avoid using multiple mocks in tests whenever possible.

**PRO**

- allows defining implementation details expectations towards a dependency
- flexible to do many things
- almost every person familiar with using this testing double

**CON**

- everything that applies to stub and spy
- introduce technical debt in the project's testing suite
- have a high risk of misusing it, and make your test focus on implementation details.
- if avoiding mocks is not an option, that's possibly feedback about the project software design state.

**Use**

- develop components in parallel with large or distributed teams.
- fault injection
- allows isolated unit testing in projects with high entropy level

**Example**

```go
m := mocks.NewMockXY(ctrl)
m.EXPECT().MyMethodName(gomock.Any()).Do(func(f func()) { f() }).AnyTimes()
```

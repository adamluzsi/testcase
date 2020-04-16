<!-- START doctoc generated TOC please keep comment here to allow auto update -->
<!-- DON'T EDIT THIS SECTION, INSTEAD RE-RUN doctoc TO UPDATE -->
**Table of Contents**  *generated with [DocToc](https://github.com/thlorenz/doctoc)*

- [Behavior Driven Development](#behavior-driven-development)
  - [So, what is BDD testing?](#so-what-is-bdd-testing)
    - [Example](#example)
  - [What is not BDD?](#what-is-not-bdd)

<!-- END doctoc generated TOC please keep comment here to allow auto update -->

# Behavior Driven Development

## So, what is BDD testing?

Testing is a methodology where code being tested.

Automatic testing is based on top of this by describing the practice to create tests that can run automatically.
This helps to ensure that a developer doesn't have to know inside a project,
before they can apply changes to the software.

[TDD][tdd-wiki-link] is extending this by defining **when** that test should be written.
By giving a set of practices such as the RED/GREEN/REFACTOR 
or the pair-programming ping pong it helps to shift the mindset 
only to write code that is specified by some form of automatic testing 

On top of [TDD][tdd-wiki-link], there is [BDD][bdd-wiki-link] which extends further the testing practices by 
requesting **discipline** on **what** you test when you describe a code specification.
In a simplified summary, one of the most known [BDD][bdd-wiki-link] testing principles is
to test the given code in the subject's **behavior** instead of the implementation.
Practicing this discipline can help to identify architecture smells in whichever place you find out that
[BDD][bdd-wiki-link] principles are violated.

This testing approach is right to any of the architecture layers
and not restricted to a particular type of testing category.
It can be applied in unit testing, integration testing, and E2E acceptance testing as well. 

One of the key challenges with this is that you are forced by BDD principles to use as real as possible testing contexts
when you were creating your test expectations and scenarios.   
The benefit of doing this is that you can freely upgrade/swag the underlying dependencies of a given code,
and as long it works with the current implementation, it is safe to be used.
This allows a faster feedback loop instead of seeing problems in staging environments like with manual testing. 
Also a benefit of keeping things as real as possible is that essentially integration tests
with domain business rule objects become really easy and convenient, and yields in integration testing practices as well.

Another aspect of the BDD testing is that unless the system behavior changes on an already defined part,
tests should not change when code is being refactored.
Creating the tests in a way that they purely focus on the behavior of the given subject should yield stable and maintainable tests.
If this is not a case, it is usually a testing smell. 

### Example

To give an example to this, if you work on a supplier implementation in an external resource layer,
namely, a structure that works with a database and you find tests that use mocking to assert input SQL's,
then you can be sure you found a violation.
It is not a behavior that a given SQL string in this case, but implementation detail.
What is behavior is what is an expected result of a given role function, in different runtime context.

For e.g.:
    your test subject is a function that returns resources from the DB in an aggregated way.
    In case you have specific values in the resource fact table, then you expect a specific result.
    In case you don't have values in the resource fact table, then you expect an empty result.
    And so on.

To follow the BDD principle of test with as close to reality as possible in this case would force the developer
to have a real database running and test against that database, the supplier implementation that works with the database.    

By following this rule, it a low hanging fruit to test a new version of the database with the same testing suite,
or to completely rewrite the current implementation for maintainability or performance reasons.

Maybe a different query can bring a huge boost, while none of your tests must change in the process of refactoring.
One thing sure, you can freely focus on the code while no system behavior change is expected.  

## What is not BDD?

When you the usage of `Given`, `When`, `And`, `Then` keyword. 
These keywords often used for describing a scenario in a BDD testing style,
but they are not the main point of BDD testing.

A usual testing smell with this is when a testing tool is abused
and it becomes hard to express and maintain tests with it.

For example, if you or your team use [cucumber][cucumber-link] for writing integration tests,
it does not necessarily mean you are doing BDD or you doing BDD correctly.
In this concrete example if you saw a team using [cucumber][cucumber-link] in a way that
developers write [cucumber][cucumber-link] tests for themselves,
then you might think about that what's the point of a tool that was meant to build a bridge between business experts
and R&D if only R&D people use it for themselves.
Why not use something that was made with R&D needs in mind then?

[tdd-wiki-link]: https://en.wikipedia.org/wiki/Test-driven_development
[bdd-wiki-link]: https://en.wikipedia.org/wiki/Behavior-driven_development
[cucumber-link]: https://cucumber.io/

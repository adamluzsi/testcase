<!-- START doctoc generated TOC please keep comment here to allow auto update -->
<!-- DON'T EDIT THIS SECTION, INSTEAD RE-RUN doctoc TO UPDATE -->
**Table of Contents**  *generated with [DocToc](https://github.com/thlorenz/doctoc)*

- [`testcase` testing Guide](#testcase-testing-guide)
  - [Why Test?](#why-test)
  - [Index](#index)
  - [Official API Documentation](#official-api-documentation)
  - [Additional Topics](#additional-topics)

<!-- END doctoc generated TOC please keep comment here to allow auto update -->

Work In Progress, the below written document is in a early draft state.

# `testcase` testing Guide

## Why Test?

The primary goal of testing is increase the speed at you gain feedback on the current system design. 
Usually when something is difficult to test, it is a good sign,
that the design might suffer from coupling, violates design principles.
In the below guide, each convention of the framework will be explained what design problem it tries to indicate.

The secondary goal of testing is to decouple the workflow into smaller units,
thus increase the efficiency by reducing the need for high mental model capacity.
As a bonus, working with tests allow you to be more resilient against random interruptions,
or get back faster to an older code base which you didn't touch since long time if ever.      

As a side effect of these conventions, the project naturally gain greater architecture flexibility and maintainability.
The lack of these aspects often credited for projects labelled as "legacy".

## Index

- [`testcase`'s testing conventions guide](/docs/guide/conventions.md) [WIP-70%]
- [TDD contracts, aka role interface testing suites](/docs/guide/contracts.md) [WIP-1%]
- [design system with DRY TDD for long term maintainability wins](/docs/guide/spechelper.md) [WIP-1%]

## Official API Documentation

If you already use the framework, and you just want pick an example,
you can go directly to the API documentation that is kept in godoc format.
- [godoc](https://godoc.org/github.com/adamluzsi/testcase)
- [pkg.go.dev](https://pkg.go.dev/github.com/adamluzsi/testcase).

## Additional Topics
- [What is BDD, and what benefits it can bring to me?](/docs/bdd.md)
- [how to define and use spec variables in testcase](/docs/variables.md)
- [high level overview of the Spec structure](/docs/spec)
- [nesting guide](/docs/nesting.md)
- [To what problem, this project give a solution?](/docs/why.md)
- [Case Study Of The Package Origin](/docs/history.md)
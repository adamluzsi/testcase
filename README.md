<!-- START doctoc generated TOC please keep comment here to allow auto update -->
<!-- DON'T EDIT THIS SECTION, INSTEAD RE-RUN doctoc TO UPDATE -->
**Table of Contents**

- [testcase](#testcase)
  - [My totally Biased Opinion about this project](#my-totally-biased-opinion-about-this-project)
  - [How much this project will be maintained ?](#how-much-this-project-will-be-maintained-)
  - [The reason behind the package](#the-reason-behind-the-package)
  - [What makes testcase different ?](#what-makes-testcase-different-)
  - [Reference Project](#reference-project)
  - [Documentations](#documentations)

<!-- END doctoc generated TOC please keep comment here to allow auto update -->

[![Mentioned in Awesome Go](https://awesome.re/mentioned-badge.svg)](https://github.com/avelino/awesome-go)
[![GoDoc](https://godoc.org/github.com/adamluzsi/testcase?status.png)](https://godoc.org/github.com/adamluzsi/testcase)
[![Build Status](https://travis-ci.org/adamluzsi/testcase.svg?branch=master)](https://travis-ci.org/adamluzsi/testcase)
[![Go Report Card](https://goreportcard.com/badge/github.com/adamluzsi/testcase)](https://goreportcard.com/report/github.com/adamluzsi/testcase)
[![codecov](https://codecov.io/gh/adamluzsi/testcase/branch/master/graph/badge.svg)](https://codecov.io/gh/adamluzsi/testcase)
# testcase

The package considered stable and no changes expected to the package exported API.

The main documentation is kept in the [GoDoc](https://godoc.org/github.com/adamluzsi/testcase),
and this README serves only as a high level introduction.

This package implements two approaches to help you to do nested BDD style testing in golang.

The package may seems inactive maybe, but it is used daily,
I just don't plan to feature creep it,
because it is totally efficient to achieve what I need,
and If I need extra helper function or anything like that,
I usually put it under the $PROJECT_ROOT/testing package.
Then I include the helpers with `.` importing.
I highly discourage the use of the  dot notation based import outside of the testing files.

## My totally Biased Opinion about this project

Primary I made this project for myself,
because using vanilla`testing#T.Run` forced me to apply repetitive boilerplate
in every test, and I wanted to introduce some form of maintainability for my tests.
I want to stick as much as possible with the core testing pkg,
so this mainly just to have those boilerplate in the form of centralized package.

I normally okay with my creations,
but I really really love this project,
because it give me a huge productivity boost,
and also it helps to apply my convention for testing.
It may not for everyone, and that is totally fine.
There are tons of testing frameworks out there,
with huge community support.

Also I need to mention, that this project is heavily based on the experience I made working with [rspec](https://github.com/rspec/rspec).
I highly recommend checking out that project and the [community takeaways about how to write a better software specification](http://www.betterspecs.org).

I don't plan on doing complex custom things in this package.
For example I don't plan to have a visually appealing reporting output
or custom assertion helpers.
No, kind the opposite, since the output intentionally looks like vanilla `testing` run output.
I need the ability to keep things close to core go testing pkg conventions,
so I can use things like `-run 'rgx'` flag.

Therefore again this project is here for my own work primary,
but please feel free to use it if you see value in it for yourself.

The project only goal is to make it easy and productive to create isolated test cases,
reproducible setup/teardown logic
and testing context based variable scoping.

## How much this project will be maintained ?

This project is based on the `testing` package [T.Run](https://godoc.org/testing#T.Run) *idiom*,
so basically as long that is supported and maintained by the golang core team,
this project is easily considered up to date.

I use it for my private projects,
but I designed this project to be cost effective for my time.
I only piggybacking the core golang team work basically.

## [The reason behind the package](https://godoc.org/github.com/adamluzsi/testcase#hdr-The_reason_behind_the_package)
## [What makes testcase different ?](https://godoc.org/github.com/adamluzsi/testcase#hdr-What_makes_testcase_different)

## Reference Project
* [toggler](https://github.com/adamluzsi/toggler)

## Documentations
* [Spec](/docs/spec/README.md)
* [Steps](/docs/steps/README.md)
* [Nesting guide](/docs/nesting.md)

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

## [To what problem, this project give a solution? (link)](/docs/why.md)

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

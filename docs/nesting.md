<!-- START doctoc generated TOC please keep comment here to allow auto update -->
<!-- DON'T EDIT THIS SECTION, INSTEAD RE-RUN doctoc TO UPDATE -->
**Table of Contents**  *generated with [DocToc](https://github.com/thlorenz/doctoc)*

- [testcase nesting guide](#testcase-nesting-guide)

<!-- END doctoc generated TOC please keep comment here to allow auto update -->

# testcase nesting guide

In `testcase` to express certain edge cases,
the framework prefers the usage of nesting.

By convention every `if` statement should have 2 corresponding testing Context.
This is required in order to keep clean track of the code complexity.
If the test coverage became "too nested", 
that is the clear sign that the implementation has too broad scope,
and the code complexity/readability also likely to be affected by the current design.

* [example code](/examples/ValidateName.go)
* [example test](/examples/ValidateName_test.go)

There is one exception from this with `testcase#Steps`,
and that is error handling.
If a implementation in subject is depending on a component that can result in error,
the specification don't have to enforce the rainy paths as a context.
Due to the nature of `testcase#Steps`, the testing suite cannot be executed concurrently,
so it is safe to use less strict isolation with that.

In situations where you have to integrate multiple component in an `interactor`,
`testcase#Steps` is preferred because it allow flat test suites.
For implementations where you need to test business logic, 
`testcase#Spec` is suggested, even if the spec has too many nested layers.
That is only represent the complexity of the component.

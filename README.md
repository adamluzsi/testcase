# testrunctx

This package implements two approach to help you do nested BDD style testing in golang with testing.T#Run function.
One is to depend on the defer func and modify a *CTX object,
and in each edge case setup the environment for the given test case,
and the another is a variable shadowing scope based approach,
where you define setup steps in each nesting level,
and the function variable scope will clean up your context after the function lifecycle ends.

For examples, test out the *_test file contents.

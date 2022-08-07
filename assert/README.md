<!-- START doctoc generated TOC please keep comment here to allow auto update -->
<!-- DON'T EDIT THIS SECTION, INSTEAD RE-RUN doctoc TO UPDATE -->

- [testcase/assert](#testcaseassert)

<!-- END doctoc generated TOC please keep comment here to allow auto update -->

# testcase/assert

This package meant to provide a small and dependency lightweight implementation for common assertion related
requirements.

- [Go pkg documentation](https://pkg.go.dev/github.com/adamluzsi/testcase/assert)

Example:

```go
assert.Should(tb).True(true)
assert.Must(tb).Equal(expected, actual, "error message")
assert.Must(tb).NotEqual(true, false, "exp")
assert.Must(tb).Contain([]int{1, 2, 3}, 3, "exp")
assert.Must(tb).Contain([]int{1, 2, 3}, []int{1, 2}, "exp")
assert.Must(tb).Contain(map[string]int{"The Answer": 42, "oth": 13}, map[string]int{"The Answer": 42}, "exp")
```

For more examples, check out the [example_test.go](./example_test.go) file.
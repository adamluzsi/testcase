package assert

import "github.com/adamluzsi/testcase/pp"

type diffFn func(value, othValue any) string

// DiffFunc is the function that will be used to print out two object if they are not equal.
// You can use your preferred diff implementation if you are not happy with the pretty print diff format.
var DiffFunc diffFn = pp.DiffFormat[any]

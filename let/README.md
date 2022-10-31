# Package let

Package `let` contains Common Testcase variable `#Let` declarations for testing purpose.

```go
var (
	ctx  = let.Context(s)
	name = let.FirstName(s)
)
act := func(t *testcase.T) error {
	return MyFunc(ctx.Get(t), name.Get(t))
}
```
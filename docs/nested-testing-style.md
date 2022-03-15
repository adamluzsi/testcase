<!-- START doctoc generated TOC please keep comment here to allow auto update -->
<!-- DON'T EDIT THIS SECTION, INSTEAD RE-RUN doctoc TO UPDATE -->


- [Nested testing style](#nested-testing-style)
  - [Power of Two](#power-of-two)
  - [Spec definition scope VS Test execution scope](#spec-definition-scope-vs-test-execution-scope)
    - [Test Context Specification Scope](#test-context-specification-scope)
    - [Test Runtime Scope](#test-runtime-scope)
  - [Describe + Immutable Subject to express `Act`](#describe--immutable-subject-to-express-act)
  - [Testing `Arrange` Hooks for DRY testing paths](#testing-arrange-hooks-for-dry-testing-paths)
  - [Don't depend on test case execution order.](#dont-depend-on-test-case-execution-order)
  - [Extendability of the testing suite](#extendability-of-the-testing-suite)
  - [Flattening nested tests](#flattening-nested-tests)

<!-- END doctoc generated TOC please keep comment here to allow auto update -->

# Nested testing style

`testcase` aim to utilize a technique called nested testing style.
In this section it will be described what reasons lead to use nested testing style in `testcase`,
along with the pros and cons of each convention. 

While nested testing style has a steep learning curve, similar to what first time [vim](https://www.vim.org/) users experience.
The guide here aims to help to explain the subtle aspects of this approach.

Nested testing style is not for everyone.
Without first investing time and effort into learning the underlying principles that explains the aspect of nested testing style,
it might give unfamiliarity and displeasing feeling to many.

After understanding the principles behind nested testing style, 
the steep learning curve will flatten out,
and the productivity will increase.

## Power of Two

One of the benefit when you use nested testing style
is that you get a visual feedback about your test subject's code complexity.
If there are way too many layer of nesting, your code likely to have high complexity to read.
Often such complexity is difficult to be spotted by a fresh reader
who may lack the contextual domain knowledge of this code piece. 

For example, each `if` statement splits the code path into two direction.
This is represented in nested testing style
by defining two different context to express the behavioral requirements for the two code path.

This nesting helps visualise and document the reason why a given complexity of `power of two` is required.
If you can't express a behavior based reason two justify the complexity,
often those `if` statements can be refactored out from the code and solved with a different idiom.

Too many `if` statement together in the form of guard clauses might hide the actual mental model capacity need from
you to interpret the code as it will execute, 
and then it becomes easier to slip and introduce a bug in one of the edge case of the given code.

[Example](/docs/examples/if_test.go)

```
if condition {
  // -> A
} else {
  // -> B
}
```

## Spec definition scope VS Test execution scope

### Test Context Specification Scope

When you write with nested testing style, you must be aware
that there is a strict differentiation between
test context specification scope and test runtime scope.

In the test context specification, you should only focus defining the context of a given testing runtime scope,
by documenting the context of a certain edge case, 
the expected behavior of the test
and bind values with `testcase.Var`iables in a given testing scope.

testing contexts are a powerful way to make your tests clear
and well organized (they keep the assertion layer easy to read).

When describing a context, start its description with 'when', 'with' or 'without'
or use the DSL functions
[`Spec#Describe`](https://pkg.go.dev/github.com/adamluzsi/testcase#Spec.Describe),
[`Spec#When`](https://pkg.go.dev/github.com/adamluzsi/testcase#Spec.When),
[`Spec#And`](https://pkg.go.dev/github.com/adamluzsi/testcase#Spec.And).

You can define facts by assigning values to test variables 
that visible from a certain testing context's scope
and below in the sub-contexts.

`testcase` don't use zero values in testing variables, instead you have to explicitly describe values,
and they behavioral reasons in a given testing scope.

```
// this is the test context specification scope
s := testcase.NewSpec(tb)

// you may document here while you build test execution context
s.Context(`documentation text here`, func(s *testcase.Spec){})

// or define test variables which are stateless outside of a test runtime execution
input := testcase.Var{ID: "I only able to fetch state during test execution"}

// but you should avoid to set dynamic values outside of the testing scope
val := &MyStruct{Config: "Value"}
var counter int
``` 

The reason you should avoid setting testing related inputs values in the test context specification scope,
is because those values will leak across test executions,
and potentially build implicit dependency on test execution order.

To set values to each test please consider using one of the two option:
- [#Let](https://pkg.go.dev/github.com/adamluzsi/testcase#Var.Let)
    * [example](https://pkg.go.dev/github.com/adamluzsi/testcase#example-Var.Let)
- [#LetValue](https://pkg.go.dev/github.com/adamluzsi/testcase#Var.LetValue)
    * [example](https://pkg.go.dev/github.com/adamluzsi/testcase#example-Var.LetValue)

This approach provides the benefit that variables isolated and only visible to they own test runtime context,
As a bonus to this discipline, if your test don't works with side effects (globals, external resource states, etc)
then you can flag the test with [`NoSideEffect`](https://pkg.go.dev/github.com/adamluzsi/testcase#Spec.NoSideEffect)
and it will run test cases concurrently for the same testing subject.  

There is a few notable exception to 
when it is acceptable to set test runtime value at test context specification scope level,
and those are constant values and immutable values which can be easily set using `#LetValue`.

```
s := testcase.NewSpec(tb)
testRuntimeVariable := testcase.Var{ID: "test runtime variable"} // T<int>

// This #LetValue will not affect the test variable in the test context spec scope,
// but will bind a value to the current *Spec for the given variable,
// which will be available only during the test runtime. 
testRuntimeVariable.LetValue(s, 42)

s.Test(``, func(t *testcase.T) {
	testRuntimeVariable.Get(t) // == 42
})
```

The other possible exception for testing suite optimization
when you need a shared resource connection injected into many test.
This shared resource connection should provide isolation (transactions) between tests runs to be safe to use.
You can manage the lifecycle of the isolation through defining Arrange and Teardown with a `testcase.Var`.
[Example to shared resource in specs](/docs/examples/spechelper_sharedResource_test.go)

### Test Runtime Scope

If test context specification scope all about defining basic facts and explanations about testing contexts,
then test runtime scope is all about what happens during test execution.
It includes the execution of all dynamic [`arranges`](/docs/aaa.md),
the [`act`](/docs/aaa.md) itself
and the [`assertions`](/docs/aaa.md).

```
s := testcase.NewSpec(tb)

testRuntimeVariable := testcase.Var{ID: "test runtime variable"}

testRuntimeVariable.Let(s, func(t *testcase.T) interface{} {
	// test runtime scope here
	return 42
})

s.Before(func(t *testcase.T) {
	// test runtime scope here
})

s.After(func(t *testcase.T) {
	// test runtime scope here
})

s.Around(func(t *testcase.T) func() {
	// test runtime scope here
	return func() { /* test runtime scope here */ }
})

s.Test(``, func(t *testcase.T) {
	// test runtime scope here
})
```

The test runtimes isolated from each other, 
by default every `testcase.Var` value only visible from and can be accessed from the given test runtime that is running.

If the test context don't utilize global/shared values which are not isolated to per test execution,
and any mutable value that used during the test managed with `testcase.Var`s,
then your test can be considered safe for parallel execution.   
This should give a gentle speed bonus to keep local development feedback loop nimble.

If you know that your test subject has no side effect,
you can flag the current test context specification scope with 
[`Spec#NoSideEffect`](https://pkg.go.dev/github.com/adamluzsi/testcase#Spec.NoSideEffect). 

## Describe + Immutable Subject to express [`Act`](/docs/aaa.md)

`testcase` suggest you that each time you when you write a test, 
make sure, that it is clear what is the testing subject.
The convention to do so is by opening [`Spec#Describe`](https://pkg.go.dev/github.com/adamluzsi/testcase#Spec.Describe) scope,
and then defining a function that will represent the [`act`](/docs/aaa.md) of the described tests.

Name this function as `subject` or as the action it meant to express.
The `subject` function input should be a `*testcase.T` and the output signature should match the [`act`](/docs/aaa.md) output.
If the `subject` of tests within the describe-block is a method,
then the `subject` return signature should match the method's output signature.
This approach with the testing `subject` should allow you to create a immutable and DRY [`act`](/docs/aaa.md).
`subject` will help with maintenance and cleanness aspects of the test coverage.

If the `subject` function content requires inputs to execute [`act`](/docs/aaa.md),
then use `testcase.Var`s as placeholders for the inputs,
and access the `Var` content through [`Var#Get`](https://pkg.go.dev/github.com/adamluzsi/testcase#Var.Get).
This allows you to define test subject without any input defined at the describe-block level context scope.
Each time you need to concretise the `testcase.Var` input for the subject,
open a new sub `Spec#Context`, describe the behavioral aspect of the value that you need to assign to the `testcase.Var`,
and then use [`Var#Let`](https://pkg.go.dev/github.com/adamluzsi/testcase#Var.Let) (or #LetValue) to assign value in that scope.
This approach ensures that even if you forgot to set a value, the framework will remind you about values you forgot to describe.

[Example](/docs/examples/immutableAct_test.go)

```
s.Describe(`#Shrug`, func(s *testcase.Spec) {
	var (
		message    = testcase.LetValue(s, fixtures.Random.String())
		subject    = func(t *testcase.T) string {
			return myStructGet(t).Shrug(message.Get(t))
		}
	)

	// ... context building (Arrange) and then Assertions.
})
```

Whenever you want to affect the subject you either need to affect the dependencies/arguments
which it uses through `testcase.Var`iables.

The benefit of this approach is to ensure that the test subject is always used in the same way,
and no accidentally configured inputs is provided without reasoning about the need.
When someone comes to refactor the code base, a failing test would clearly describe what is the test edge case context
that lead to the usage of the test subject that broke during the test.

And lastly but not least, the main goal with this to unify the test execution flow,
and thus reduce the required mental model to understand the test at a given context.
Your brain can instantly rely on the fact that the subject will never change,
and only the context that changes.

## Testing [`Arrange`](/docs/aaa.md) Hooks for DRY testing paths

When you describe a common testing edge case where similar contextual arranges present for test cases,
you can use combine `Spec#Context` with `Spec#Before` to express this.

In a simplified example for using Hooks, you can simplify from this:

```
t.Run(``, func(t *testing.T) {
    t.Log("foo")
    t.Cleanup(func() { t.Log("bar") })
    // act
    // assert
})

t.Run(``, func(t *testing.T) {
    t.Log("foo")
    t.Cleanup(func() { t.Log("bar") })
    // act
    // assert
})

t.Run(``, func(t *testing.T) {
    t.Log("foo")
    t.Cleanup(func() { t.Log("bar") })
    // act
    // assert
})
```

to

```
s := testcase.NewSpec(tb)

s.Before(func(t *testcase.T) {
    t.Log("foo")
})

s.After(func(t *testcase.T) {
    t.Log("bar")
})

s.Test(``, func(t *testcase.T) {
    // act + assert
})

s.Test(``, func(t *testcase.T) {
    // act + assert
})

s.Test(``, func(t *testcase.T) {
    // act + assert
})
```

In this example where we only do some logging,
the impact of using hooks might not be as obvious,
but as your testing suite requires to describe more and more behavioral edge cases,
the arrange and cleanup becomes more repetitive.
This also introduce difficulty in maintenance, 
by forcing the developer to creat test helper functions.

In `testcase` this comes naturally with the framework usage,
and also with the `testcase#Var`iables,
where setup and cleanup becomes part of the testing suite by using a variable.

Spec Hooks express test runtime scope, 
and should not manage non isolated resources
from the test context specification scope.

## Don't depend on test case execution order.

Your test should avoid depending on the order of the execution of individual test cases. 

But why?

Have you ever seen a unit test pass in isolation, but fail when run in a suite?
Or vice versa, pass in a suite, but fail when run in isolation?
What drives some of us to do this in the first place?

The most common case is when the first test performs some action which results in side effect.
The temptation might be strong to use this side effect as the starting point for the next test.
While the whole testing suite is beign executed, in a certain order,
the test execution order dependency will remain hidden for the next developer. 

A testing suite is also something that evolves with a project.
New tests will be added, and old tests will be deleted,
and some will be updated to express changes in the business rules.
To avoid problems as our test suites grow and change,
it's important to keep test cases independent.

In `testcase` conventions, whenever you need to depend on a side effect,
you should express it clearly with a combination of
[`Spec#Context`](https://pkg.go.dev/github.com/adamluzsi/testcase#Spec.Context)
where you document the event that caused the side effect
and within that context, you should execute the event in a 
[`Spec#Before`](https://pkg.go.dev/github.com/adamluzsi/testcase#Spec.Before) 
or [`Spec#Around`](https://pkg.go.dev/github.com/adamluzsi/testcase#Spec.Around) block.  

This approach ensures that each test documents and arrange its requirements.
There are actual events to arranged to a given testing scope, which will be executed before an act or assertions.  

To sum this up,

> They must not depend upon any assumed initial state,
  they must not leave any residue behind that would prevent them from being re-run.

The `testcase` framework will shuffle the execution order of your specification, 
thus potentially reveal ordering dependencies in your testing suite. 

```
// BAD example:

s := testcase.NewSpec(tb)

s.Test(``, func(t *testcase.T) { /* create entity in a external resource */ })
s.Test(``, func(t *testcase.T) { /* use the entity created in the external resource from the previous test /* })
```

```
// GOOD example:

s := testcase.NewSpec(tb)

s.Test(``, func(t *testcase.T) { /* create entity in a external resource */ })

s.When(`xy present in the storage`, func(s *testcase.Spec) {
  s.Before(func(t *testcase.T) { /* create entity in a external resource */ })

  s.Test(``, func(t *testcase.T) { /* use the entity created in the external resource /* })
})
```

## Extendability of the testing suite

You can describe business rule requirement as a series of testing context arrange.
If you structure your testing suite through using `Spec#Context`.

This way, if a business requirement changes for a certain edge context,
it should be obvious where to apply changes 
or where to extend the testing suite with further assertion as test cases.

## Flattening nested tests

While the nested testing style has benefits, for some reader,
it can be challenging to read a test if it has way too many levels of nesting.
For that in testcase, we suggest grouping testing context branches that have a common goal.
If you create a top-level function which takes `*testing.Spec` as the first parameter,
then you can move a part of the testing specification under that top-level function.

This technique can be used to flatten tests of
- per endpoint REST handler tests
- per method Struct tests
- shared specifications
    * like common test cases which would otherwise repeat between testing contexts

```
var Example = testcase.Var{
    ID: "Example",
    Init: func(t *testcase.T) interface{} {
        return &mypkg.Example{} 
    }
}

func TestExample(t *testing.T) {
    s := testcase.NewSpec(t)
    Example.Let(s, nil)
    s.Before(func(t *testcase.T) { /* common setup */ })
    s.Describe(".Something", SpecExampleSomething)
    // ... 
}

func SpecTSomething(s *testcase.Spec) {
    subject := func(t *testcase.T) error {
        return Example.Get(t).Something()       
    }

    // ...
}
```
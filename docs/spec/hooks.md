<!-- START doctoc generated TOC please keep comment here to allow auto update -->
<!-- DON'T EDIT THIS SECTION, INSTEAD RE-RUN doctoc TO UPDATE -->


- [Spec Hooks](#spec-hooks)
  - [Before](#before)
  - [After](#after)
  - [Around](#around)

<!-- END doctoc generated TOC please keep comment here to allow auto update -->

# Spec Hooks

Hooks help you setup common things for each test case.
For example clean ahead, clean up, mock expectation configuration,
and similar things can be done in hooks,
so your test case blocks with `Then` only represent the expected result(s).

In case you work with something that depends on side-effects,
such as database tests, you can use the hooks,
to create clean-ahead / clean-up blocks.

Also if you use gomock, you can use the spec#Around function,
to set up the mock with a controller, and in the teardown function,
call the gomock.Controller#Finish function,
so your test cases will be only about
what is the different behavior from the rest of the test cases.

It will panic if you use hooks or variable preparation in an ambiguous way,
or when you try to access variable that doesn't exist in the context where you do so.
It tries to panic with friendly and supportive messages, but that is highly subjective.

## Before

Before give you the ability to run a block before each test case.
This is ideal for doing clean ahead before each test case.
The received *testing.T object is the same as the Then block *testing.T object
This hook applied to this scope and anything that is nested from here.
All setup block is stackable.

```go
s := testcase.NewSpec(t)

s.Before(func(t *testcase.T) {
    // this will run before the test cases.
})
```

## After

After give you the ability to run a block after each test case.
This is ideal for running cleanups.
The received *testing.T object is the same as the Then block *testing.T object
This hook applied to this scope and anything that is nested from here.
All setup block is stackable.

```go
s := testcase.NewSpec(t)

s.After(func(t *testcase.T) {
    // this will run after the test cases.
    // this hook applied to this scope and anything that is nested from here.
    // hooks can be stacked with each call.
})
```

## Around

Around give you the ability to create "Before" setup for each test case,
with the additional ability that the returned function will be deferred to run after the Then block is done.
This is ideal for setting up mocks, and then return the assertion request calls in the return func.
This hook applied to this scope and anything that is nested from here.
All setup block is stackable.

```go
s := testcase.NewSpec(t)

s.Around(func(t *testcase.T) func() {
    // this will run before the test cases

    // this hook applied to this scope and anything that is nested from here.
    // hooks can be stacked with each call
    return func() {
        // The content of the returned func will be deferred to run after the test cases.
    }
})
```

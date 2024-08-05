<!-- START doctoc generated TOC please keep comment here to allow auto update -->
<!-- DON'T EDIT THIS SECTION, INSTEAD RE-RUN doctoc TO UPDATE -->

- [Clock and Timecop](#clock-and-timecop)
  - [INSTALL](#install)
  - [FEATURES](#features)
  - [Freezing time](#freezing-time)
  - [USAGE](#usage)
    - [timecop.Travel + timecop.Freeze](#timecoptravel--timecopfreeze)
    - [timecop.SetSpeed](#timecopsetspeed)
  - [Design](#design)
  - [References](#references)
  - [FAQ](#faq)
    - [What is the performance impact on my production code if I use clock?](#what-is-the-performance-impact-on-my-production-code-if-i-use-clock)
    - [Why not pass a function argument or time value directly to a function/method?](#why-not-pass-a-function-argument-or-time-value-directly-to-a-functionmethod)
    - [Will this replace dependency injection for time-related configurations?](#will-this-replace-dependency-injection-for-time-related-configurations)
    - [Why not just use a global variable with `time.Now`?](#why-not-just-use-a-global-variable-with-timenow)
    - [How does the clock package help with testing time-dependent goroutine synchronization?](#how-does-the-clock-package-help-with-testing-time-dependent-goroutine-synchronization)
    - [Does the `clock` package have a monotonic view of time?](#does-the-clock-package-have-a-monotonic-view-of-time)

<!-- END doctoc generated TOC please keep comment here to allow auto update -->

# Clock and Timecop

The `clock` package provides advanced "time travel" and "time scaling" features for easy testing of time-dependent code. 
As a drop-in replacement for the standard `time` package, it simplifies simulating and manipulating time in your applications, ensuring smooth integration and improved functionality for your time-based tests.

## INSTALL

```sh
go get go.llib.dev/testcase/clock
```

## FEATURES

- Drop in replacement for standard `time` package
- Freeze time to a specific point.
- Travel back to a specific time, but allow time to continue moving forward.
- Scale time by a given scaling factor will cause the time to move at an accelerated pace.
- No dependencies other than the stdlib
- Continous and nested calls with timecop.Travel are supported
- Works with any regular Go projects

## Freezing time

Using the `timecop` package, you can freeze time in `clock` at different levels during a time travelling.

The first level, `timecop.Freeze`, stops the timeline from moving forward but doesn't halt tickers or functions that depend on time. This is useful for testing components that depends on background goroutines that has time-based scheduling, allowing events to fire while ensuring you can assert time values in entity fields.
Think of it like freezing a river — the surface is still, but the flow underneath continues.

The second level, `timecop.DeepFreeze`, completely freezes all time-related values and prevents functions like `clock.Ticker`, `clock.After`, and `clock.Sleep` from firing until time is moved forward. This is useful for testing components with time-sensitive behaviour.
With the river example, deep freeze would mean turning a river into something like a glacier.

## USAGE

```go
package main

import (
   "go.llib.dev/testcase/assert"
   "go.llib.dev/testcase/clock"
   "go.llib.dev/testcase/clock/timecop"
   "testing"
   "time"
)

func Test(t *testing.T) {
   type Entity struct {
      CreatedAt time.Time
   }

   MyFunc := func() Entity {
      return Entity{
         CreatedAt: clock.TimeNow(),
      }
   }

   expected := Entity{
      CreatedAt: clock.TimeNow(),
   }

   timecop.Travel(t, expected.CreatedAt, timecop.Freeze)

   assert.Equal(t, expected, MyFunc())
}
```

Time travelling is undone as part of the test's teardown.

### timecop.Travel + timecop.Freeze

The Freeze option causes the observed time to stop until the first time reading event.

### timecop.SetSpeed

Let's say you want to test a "live" integration wherein entire days could pass by
in minutes while you're able to simulate "real" activity. For example, one such use case
is being able to test reports and invoices that run in 30-day cycles in very little time while also
simulating activity via subsequent calls to your application.

```go
timecop.SetSpeed(t, 1000) // accelerate speed by 1000x times from now on. 
<-clock.After(time.Hour) // takes only 1/1000 time to finish, not an hour.
clock.Sleep(time.Hour) // same
```

## Design

The package uses a singleton pattern.
The original design had a Clock and a Timecop type to do dependency injection,
but upon doing spiking with it, it felt foreign to how we currently use time.Now() or time.After(duration).
Also, it made it possible that different components reside in different timelines,
while time should be observed as a singleton entity by the whole application.
Time manipulation seems to be a good use case where the singleton pattern is the least wrong solution.

## References

The package was inspired by [travisjeffery' timecop project](https://github.com/travisjeffery/timecop).

## FAQ

### What is the performance impact on my production code if I use clock?

There is no performance impact on your production code. 
The time-travel feature is only enabled during testing. 
In a production environment, the clock's functions act as aliases for the standard `time` functions.

### Why not pass a function argument or time value directly to a function/method?

While injecting time as an argument or dependency is a valid approach, the aim with `clock` was to keep the usage feeling idiomatic and close to the standard `time` package, while also making testing convenient and easy.

Using dependency injection for time related components may complicate high-level testing that involve many components.
In these cases, it's easier to simulate time changes with a shared `clock` package, rather than injecting the time component into all dependencies. This also allow your tests to use the same constructor functions as your production code and know little about which of its component is time sensitive. Also when components set fields like "CreatedAt" timestamps, it becomes very convinent to keep them in the same timeline, and make the assertions easy on the resulting entities.

### Will this replace dependency injection for time-related configurations?

For configurable values in your logic, you should still use dependency injection. However, you can efficiently test these configurations with the `clock` package by using time travel in your tests. For example, if you're designing a scheduler that takes `time.Duration` as a configuration input, you can freeze the time, set a specific duration in your test, inject it into your component, and then simulate different time scenarios based on the test cases you want to cover.

### Why not just use a global variable with `time.Now`?

That approach can work well for testing. If you consistently use that global variable throughout your project, it can be very helpful for integration tests. This is essentially how the `clock` library started. As more use cases emerged during our project, we expanded it to ensure testability for those scenarios too.

If you decide to use a global variable, I highly recommend creating a Stub function for it. This function should reset the value to `time.Now` after the test is done, ensuring clean tests cleanups. Also, be mindful of parallel testing and potential edge cases with it.

If you're looking to create a reusable component in the form of a shared package that supports time manipulation in tests, make sure the common package that has the time stub functionality is easily accessible to those using your package.

### How does the clock package help with testing time-dependent goroutine synchronization?

The `clock` package focuses on time manipulation rather than testing the implementation details of components using goroutines. 
It isn't designed to test time-based synchronization between goroutines, 
but rather to enable testing that focuse on the system behaviour.

For this purpose, you can use the `assert.Eventually` helper from `testcase/assert` to test systems with eventual consistency. 
This helper improves the likelihood that your goroutines will be scheduled more eagerly, 
increasing the chances that your test assertions will be met in the ideal testing time.

```go
func TestXXX(t *testing.T) {
   subject := ...

   timecop.Travel(t, time.Hour) // trigger time related behavioural change

   assert.Eventually(t, time.Second, func(t assert.It) { // assert that eventually within a second, we expect an outcome
      got := subject.LookupSomething(...)
      assert.Equal(t, got, "expected")
   })

}
```

### Does the `clock` package have a monotonic view of time?

Monotonic time means:
- Time is always increasing steadily, without jumping backwards.
- Every thread sees time moving forward at the same rate.
- All goroutines share a single, unified notion of time.

The `clock` package inherits its time view from the standard Go time API. 
The Go standard library provides APIs to get the current time, 
which includes both the "wall clock" and the "monotonic clock".
When the system clock synchronizes and changes, the wall clock time can move backward, 
but for measurements, Go uses the monotonic time value for precise results.

The `clock` package also allows traveling backward in time. 
In this sense, `clock` isn't strictly monotonic because you can set it to a specific point in time. 
After time travel, time is monotonic again until you again make a travel back in time.

It’s open to interpretation whether the `clock` remains monotonic when time is frozen with `timecop`.


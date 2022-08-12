# Clock and Timecop

## DESCRIPTION

Package providing "time travel" and "time scaling" capabilities,
making it simple to test time-dependent code.

## INSTALL

```sh
go get -u github.com/adamluzsi/testcase
```

## FEATURES

- Freeze time to a specific point.
- Travel back to a specific time, but allow time to continue moving forward.
- Scale time by a given scaling factor will cause the time to move at an accelerated pace.
- No dependencies other than the stdlib
- Nested calls to timecop.Travel is supported
- Works with any regular Go projects

## USAGE

```go
package main

import (
   "github.com/adamluzsi/testcase/assert"
   "github.com/adamluzsi/testcase/clock"
   "github.com/adamluzsi/testcase/clock/timecop"
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

   timecop.Travel(t, expected.CreatedAt, timecop.Freeze())

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

## References

The package was inspired by [travisjeffery' timecop project](https://github.com/travisjeffery/timecop).

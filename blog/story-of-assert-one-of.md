# The Story of searching for a Needle with dynamic fields within a haystack

I often work with CRUD APIs where the read operations return multiple values,
which can sometimes include creation, update or access times.
Asserting against an expected subset of these values requires more preparation than I'd prefer due to my tendency to be lazy.

---

I considered using a loop with conditional checks to trigger 'continue' when conditions weren't met,
but if I needed more than one checks, it started to become complex to read the condition checking,
and also more difficult to make relevant logs.
It also forced me two different assertion style, one with if condition based `continue` management,
and the other how I wrote everything else using `assert` helpers.

```go

var found bool
for _, v := range vs {
  if v.A != expA {
    continue
  }
  if v.B != expB {
    continue
  }

  var cInd = map[int]struct{}
  for i, gCV := range v.C {
    for _, eCV := range expC {
      if gCV == eCV {
        cInd[i] = struct{}{}
        break
      }
    }
  }
  if len(cInd) != len(v.C) {
    continue
  }
  found = true
  break
}

assert.Assert(t, found, "expected that one of the value of ... will be ...")
```

Instead of using a similar style, something like this but one that would pass if one of the value pass the matchers.

```go
for _, v := range vs {
  // incorrect pseudo code,
  // as this would fail for all,
  // and I would need only to fail if none of the value passed the assertions
  assert.Equal(t, v.A, expA)
  assert.Equal(t, v.B, expB)
  assert.ContainsExactly(t, v.C, expC)
}
```

---

To streamline this, I decided to combine these into a single assertion helper and make 'testing.TB#FailNow' act as the control to decide if 'continue' is needed.
This approach allows for clearer testing and easier debugging, while maintaining the same assertion idiom I'm already used in other scenarios.

**pseudo**:

```txt
assert that one of the slice values
    will match A expectation
    will match B expectation
    will match C expectation
```

A while ago, I needed to create an `AnyOf` assertion,
allowing multiple testing cases to be asserted, without making a test fail,
as long at least one of these testing cases match our expectations.

So if at least one scenario passes without issues, the AnyOf assertion itself will pass.
This made it into a perfect candidate to be reused to create the assertion helper I needed.

Here's a simple example where `v` variable's value is either `"foo"`, `"bar"`, or `"baz"`:  

```go
assert.AnyOf(t, func(a *assert.A) {  
  a.Case(func(t testing.TB) { assert.Equal(t, v, "foo")})  
  a.Case(func(t testing.TB) { assert.Equal(t, v, "bar")})  
  a.Case(func(t testing.TB) { assert.Equal(t, v, "baz")})  
})
```

While this could be expressed simply by using `assert.Contains` using `[]string{"foo", "bar", "baz"}`,
don't let it fool you, as it enables to test different accepted scenarios based on complex requirements.

For instance, you could test whether a struct's specific field contains a certain value or not,
then another field contains a certain value.  

It also enables assertions on individual elements of a given list value:  

```go
var vs []V
assert.AnyOf(t, func(a *A) {  
  for _, v := range vs {  
    a.Case(func(it testing.TB) {  
      // assertions on `v
    })  
    if a.OK() { // we found a winner!  
      break  
    }  
  }  
})
```

However, while the `AnyOf` tool felt as an increadibly power tool,
it also felt a bit of boilerplate to use, so it made sense to continue 
and create a dedicated `OneOf` assertion helper, by wrapping it up:

```go
assert.OneOf(t, vs, func(t testing.TB, got T) {  
  assert.Equal(t, got.V1, "The Answer")  
  assert.Equal(t, got.V2, 42)  
})
```

I found this assertion helper very useful to test struct types
where some of the fields were dynamic values such as CreatedAt, UpdatedAt, AccessedAt.

But there was an issue: logs were ignored until now. When it failed,
I found myself attempting to pretty-print the value in scenarios where several assertions had already passed.
This led me to the idea of forwarding the logs from the failed scenario that had the highest number of passing assertions,
improving my developer experience.  

Now I have the best of both worlds:
I can assess whether an expected dynamic value is part of a result set, without time manipulation, stubbing or complex arrangements,
and I still get nice error outputs in my test

Here's an example where `OneOf` helps finding the right User value in a list,
which would be painful to do using `assert.Contains` helper.

```go
type User struct {  
  ID       string  
  Username string
  Level    int

  Permissions []Permission
  // fields that could cause noise in a assert.Contains  
  CreatedAt  time.Time  
  UpdatedAt  time.Time  
  AccessedAt time.Time  
}  

var tb testing.TB  
var users []User  

assert.OneOf(tb, users, func(t testing.TB, got User) {  
  assert.Equal(t, got.Username, "expected-user-name") // one very simple assertion as a sample
  assert.Equal(t, got.Level, 42, "it was expected that after the test arrangement, the user is at lvl 42")
  assert.ContainsExactly(t, got.Permissions, expAdminPermissions)
})
```

Feel free to share any feedback or suggestions!

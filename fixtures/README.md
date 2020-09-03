# Fixtures

Fixture helps you create randomized input values.
This allows you to stress test your project with built in tools like:
> go test -count n

## Usage

### Fixture creation

```go
package main_test

import "github.com/adamluzsi/testcase/fixtures"

func Example() {
	type ExampleStruct struct {
		Name string
	}

	_ = fixtures.New(ExampleStruct{}).(*ExampleStruct)
}
```

### Random

```go
package main_test

import (
	"math/rand"
	"time"

	"github.com/adamluzsi/testcase/fixtures/random"
)

func Example() {
	rnd := random.NewRandom(rand.NewSource(time.Now().Unix()))

	_ = rnd.Bool()
	_ = rnd.Float32()
	_ = rnd.Float64()
	_ = rnd.Int()
	_ = rnd.IntBetween(24, 42)
	_ = rnd.IntN(42)
	_ = rnd.String()
	_ = rnd.StringN(42)
	_ = rnd.Time()
	_ = rnd.TimeBetween(time.Now(), time.Now().Add(time.Hour))
	_ = rnd.ElementFromSlice([]string{`foo`, `bar`, `baz`}).(string)
	_ = rnd.KeyFromMap(map[string]struct{}{
		`foo`: {},
		`bar`: {},
		`baz`: {},
	}).(string)
}
```

or as a shortcut you can use random initialized by default in the fixtures package:

```go
package main_test

import (
	"math/rand"
	"time"

	"github.com/adamluzsi/testcase/fixtures"
)

func Example() {
	_ = fixtures.Random.Bool()
	_ = fixtures.Random.Float32()
	_ = fixtures.Random.Float64()
	_ = fixtures.Random.Int()
	_ = fixtures.Random.IntBetween(24, 42)
	_ = fixtures.Random.IntN(42)
	_ = fixtures.Random.String()
	_ = fixtures.Random.StringN(42)
	_ = fixtures.Random.Time()
	_ = fixtures.Random.TimeBetween(time.Now(), time.Now().Add(time.Hour))
	_ = fixtures.Random.ElementFromSlice([]string{`foo`, `bar`, `baz`}).(string)
	_ = fixtures.Random.KeyFromMap(map[string]struct{}{
		`foo`: {},
		`bar`: {},
		`baz`: {},
	}).(string)
}
```
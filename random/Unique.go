package random

import (
	"time"

	"go.llib.dev/testcase/internal/proxy"
	"go.llib.dev/testcase/internal/reflects"
)

// Unique function is a utility that helps with generating distinct values
// from those in a given exclusion list.
// If you need multiple unique values of the same type,
// this helper function can be useful for ensuring they're all different.
//
//	rnd := random.New(random.CryptoSeed{})
//	v1 := random.Unique(rnd.Int)
//	v2 := random.Unique(rnd.Int, v1)
//	v3 := random.Unique(rnd.Int, v1, v2)
func Unique[T any](blk func() T, excludeList ...T) T {
	if len(excludeList) == 0 {
		return blk()
	}

	var (
		retries  int
		deadline = proxy.TimeNow().Add(5 * time.Second)
	)
	for ; proxy.TimeNow().Before(deadline) || retries < 5; retries++ {
		var (
			v  T    = blk()
			ok bool = true
		)
		for _, excluded := range excludeList {
			if eq(v, excluded) {
				ok = false
				break
			}
		}
		if ok {
			return v
		}
	}
	panic("random.Unique failed to find a unique value")
}

func eq[T any](a, b T) bool {
	isEqual, err := reflects.DeepEqual(a, b)
	if err != nil {
		panic(err.Error())
	}
	return isEqual
}

// UniqueValues is an option that used to express a desire for unique value generation with certain functions.
// For example if random.Slice receives the UniqueValues flag, then the created values will be guaranteed to be unique,
// unless it is not possible within a reasonable attempts using the provided value maker function.
const UniqueValues = flagUniqueValues(0)

type flagUniqueValues int

func (flagUniqueValues) sliceOption(c *sliceConfig) { c.Unique = true }
func (flagUniqueValues) mapOption(c *mapConfig)     { c.Unique = true }

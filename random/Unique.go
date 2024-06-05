package random

import (
	"time"

	"go.llib.dev/testcase/clock"
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
	deadline := clock.Now().Add(5 * time.Second)
	for clock.Now().Before(deadline) {
		var (
			v  T    = blk()
			ok bool = true
		)
		for _, excluded := range excludeList {
			isEqual, err := reflects.DeepEqual(v, excluded)
			if err != nil {
				panic(err.Error())
			}
			if isEqual {
				ok = false
			}
		}
		if ok {
			return v
		}
	}
	panic("random.Unique failed to find a unique value")
}

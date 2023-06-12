package assert

import (
	"fmt"
	"github.com/adamluzsi/testcase/assert/internal"
	"math/big"
	"net"
	"reflect"
	"testing"
	"time"
)

func eq(tb testing.TB, exp, act any) bool {
	tb.Helper()
	isEq, err := internal.DeepEqual(exp, act)
	Must(tb).NoError(err)
	return isEq
}

type EqualFunc[T any] interface {
	func(v1, v2 T) (bool, error) |
		func(v1, v2 T) bool
}

func RegisterEqual[T any, FN EqualFunc[T]](fn FN) struct{} {
	var rfn func(v1, v2 reflect.Value) (bool, error)
	switch fn := any(fn).(type) {
	case func(v1, v2 T) (bool, error):
		rfn = func(v1, v2 reflect.Value) (bool, error) {
			return fn(v1.Interface().(T), v2.Interface().(T))
		}
	case func(v1, v2 T) bool:
		rfn = func(v1, v2 reflect.Value) (bool, error) {
			return fn(v1.Interface().(T), v2.Interface().(T)), nil
		}
	default:
		panic(fmt.Sprintf("unrecognised Equality checker function signature"))
	}
	internal.RegisterIsEqual(reflect.TypeOf((*T)(nil)).Elem(), rfn)
	return struct{}{}
}

var _ = RegisterEqual[time.Time](func(t1, t2 time.Time) bool {
	return t1.Equal(t2)
})

var _ = RegisterEqual[net.IP](func(ip1, ip2 net.IP) bool {
	return ip1.Equal(ip2)
})

var _ = RegisterEqual[big.Int](func(v1, v2 big.Int) bool {
	return v1.Cmp(&v2) == v2.Cmp(&v1)
})

var _ = RegisterEqual[big.Rat](func(v1, v2 big.Rat) bool {
	return v1.Cmp(&v2) == v2.Cmp(&v1)
})

var _ = RegisterEqual[big.Float](func(v1, v2 big.Float) bool {
	return v1.Cmp(&v2) == v2.Cmp(&v1)
})

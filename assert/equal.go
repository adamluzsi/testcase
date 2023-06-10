package assert

import (
	"reflect"
	"testing"
)

func eq(tb testing.TB, exp, act any) bool {
	tb.Helper()
	
	if isEqual, ok := tryIsEqual(tb, exp, act); ok {
		return isEqual
	}

	return reflect.DeepEqual(exp, act)
}

var methodNamesForIsEqual = []string{"IsEqual", "Equal"}

func tryIsEqual(tb testing.TB, exp, act any) (isEqual bool, ok bool) {
	tb.Helper()
	defer func() { recover() }()
	expRV := reflect.ValueOf(exp)
	actRV := reflect.ValueOf(act)

	if expRV.Type() != actRV.Type() {
		return false, false
	}

	tryIsEqualMethod := func(methodName string) (bool, bool) {
		method := expRV.MethodByName(methodName)
		methodType := method.Type()

		if methodType.NumIn() != 1 {
			return false, false
		}
		if numOut := methodType.NumOut(); !(numOut == 1 || numOut == 2) {
			return false, false
		}
		if methodType.In(0) != actRV.Type() {
			return false, false
		}

		res := method.Call([]reflect.Value{actRV})

		switch {
		case methodType.NumOut() == 1: // IsEqual(T) (bool)
			return res[0].Bool(), true

		case methodType.NumOut() == 2: // IsEqual(T) (bool, error)
			Must(tb).Nil(res[1].Interface())
			return res[0].Bool(), true

		default:
			return false, false
		}
	}

	for _, methodName := range methodNamesForIsEqual {
		if eq, ok := tryIsEqualMethod(methodName); ok {
			return eq, ok
		}
	}
	return false, false
}

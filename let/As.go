package let

import (
	"fmt"
	"go.llib.dev/testcase"
	"reflect"
)

func As[To, From any](Var testcase.Var[From]) testcase.Var[To] {
	asID++
	fromType := reflect.TypeOf((*From)(nil)).Elem()
	toType := reflect.TypeOf((*To)(nil)).Elem()
	if !fromType.ConvertibleTo(toType) {
		panic(fmt.Sprintf("you can't have %s as %s", fromType.String(), toType.String()))
	}
	return testcase.Var[To]{
		ID: fmt.Sprintf("%s AS %T #%d", Var.ID, *new(To), asID),
		Init: func(t *testcase.T) To {
			var rFrom = reflect.ValueOf(Var.Get(t))
			return rFrom.Convert(toType).Interface().(To)
		},
	}
}

var asID int // adds extra safety that there won't be a name collision between two variables

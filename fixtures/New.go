package fixtures

import (
	"context"
	"reflect"
)

// New returns a populated entity for a given business data entity.
// This is primary and only used for testing.
// With fixtures, it become easy to create generic query objects
// that can be used during testing with randomly generated data.
func New[T any](opts ...Option) *T {
	var (
		ff     = &Factory{Random: Random}
		config = newConfig(opts...)
	)

	ptr := new(T)
	elem := reflect.ValueOf(ptr).Elem()

	for i := 0; i < elem.NumField(); i++ {
		v := elem.Field(i)
		sf := elem.Type().Field(i)

		if v.CanSet() && config.CanPopulateStructField(sf) {
			newValue := reflect.ValueOf(ff.Fixture(v.Interface(), context.Background()))
			if newValue.IsValid() {
				v.Set(newValue)
			}
		}
	}

	return ptr
}

package fixtures

import (
	"reflect"
)

// New returns a populated entity for a given business data entity.
// This is primary and only used for testing.
// With fixtures, it become easy to create generic query objects
// that can be used during testing with randomly generated data.
//
// DEPRECATED: please consider using fixtures.Factory instead
func New(T interface{}, opts ...Option) (pointer interface{}) {
	var (
		ff     = &Factory{Random: Random}
		config = newConfig(opts...)
	)

	ptr := reflect.New(reflect.TypeOf(T))
	elem := ptr.Elem()

	for i := 0; i < elem.NumField(); i++ {
		v := elem.Field(i)
		sf := elem.Type().Field(i)

		if v.CanSet() && config.CanPopulateStructField(sf) {
			newValue := reflect.ValueOf(ff.Create(v.Interface()))
			if newValue.IsValid() {
				v.Set(newValue)
			}
		}
	}

	return ptr.Interface()
}

package fixtures

import (
	"context"
	"math"
	"reflect"
	"sync"
	"time"

	"github.com/adamluzsi/testcase/random"
)

type Factory struct {
	Random      *random.Random
	StubContext context.Context
	Options     []Option

	config struct {
		init  sync.Once
		len   int
		value *config
	}
	types struct {
		init    sync.Once
		mapping map[reflect.Type]FactoryFunc
	}
	kinds struct {
		init    sync.Once
		mapping map[reflect.Kind]FactoryFunc
	}
	randomInit sync.Once
}

type FactoryFunc func(any) any

func (f *Factory) getConfig() *config {
	if f.config.len != len(f.Options) {
		f.config.init = sync.Once{}
	}
	f.config.init.Do(func() {
		f.config.len = len(f.Options)
		f.config.value = newConfig(f.Options...)
	})
	return f.config.value
}

func (f *Factory) getRandom() *random.Random {
	f.randomInit.Do(func() {
		if f.Random == nil {
			f.Random = Random
		}
	})
	return f.Random
}

func (f *Factory) Create(T any) any {
	if T == nil {
		// type error panic will be solved after go generics support
		panic(`nil is not accepted as input[T] type`)
	}
	rt := reflect.TypeOf(T)
	typeFunc, ok := f.getTypes()[rt]
	if ok {
		return typeFunc(T)
	}
	if kindFunc, ok := f.getKinds()[rt.Kind()]; ok {
		return kindFunc(T)
	}
	return T
}

func (f *Factory) Context() context.Context {
	if f.StubContext == nil {
		return context.Background()
	}

	return f.StubContext
}

func (f *Factory) RegisterType(T any, ff FactoryFunc) {
	f.getTypes()[reflect.TypeOf(T)] = ff
}

func (f *Factory) getTypes() map[reflect.Type]FactoryFunc {
	f.types.init.Do(func() {
		f.types.mapping = make(map[reflect.Type]FactoryFunc)
		f.types.mapping[reflect.TypeOf(int(0))] = f.int
		f.types.mapping[reflect.TypeOf(int8(0))] = f.int8
		f.types.mapping[reflect.TypeOf(int16(0))] = f.int16
		f.types.mapping[reflect.TypeOf(int32(0))] = f.int32
		f.types.mapping[reflect.TypeOf(int64(0))] = f.int64
		f.types.mapping[reflect.TypeOf(uint(0))] = f.uint
		f.types.mapping[reflect.TypeOf(uint8(0))] = f.uint8
		f.types.mapping[reflect.TypeOf(uint16(0))] = f.uint16
		f.types.mapping[reflect.TypeOf(uint32(0))] = f.uint32
		f.types.mapping[reflect.TypeOf(uint64(0))] = f.uint64
		f.types.mapping[reflect.TypeOf(float32(0))] = f.float32
		f.types.mapping[reflect.TypeOf(float64(0))] = f.float64
		f.types.mapping[reflect.TypeOf(string(""))] = f.string
		f.types.mapping[reflect.TypeOf(bool(false))] = f.bool
		f.types.mapping[reflect.TypeOf(time.Time{})] = f.timeTime
		f.types.mapping[reflect.TypeOf(time.Duration(0))] = f.timeDuration
	})
	return f.types.mapping
}

func (f *Factory) int(T any) any {
	return f.getRandom().Int()
}

func (f *Factory) int8(T any) any {
	return int8(f.getRandom().Int())
}

func (f *Factory) int16(T any) any {
	return int16(f.getRandom().Int())
}

func (f *Factory) int32(T any) any {
	return int32(f.getRandom().Int())
}

func (f *Factory) int64(T any) any {
	return int64(f.getRandom().Int())
}

func (f *Factory) uint(T any) any {
	return uint(f.getRandom().Int())
}

func (f *Factory) uint8(T any) any {
	return uint8(f.getRandom().Int())
}

func (f *Factory) uint16(T any) any {
	return uint16(f.getRandom().Int())
}

func (f *Factory) uint32(T any) any {
	return uint32(f.getRandom().Int())
}

func (f *Factory) uint64(T any) any {
	return uint64(f.getRandom().Int())
}

func (f *Factory) float32(a any) any {
	return f.getRandom().Float32()
}

func (f *Factory) float64(a any) any {
	return f.getRandom().Float64()
}

func (f *Factory) timeTime(T any) any {
	return f.getRandom().Time()
}

func (f *Factory) timeDuration(T any) any {
	return time.Duration(f.getRandom().IntBetween(int(time.Second), math.MaxInt32))
}

func (f *Factory) bool(T any) any {
	return f.getRandom().Bool()
}

func (f *Factory) string(T any) any {
	return f.getRandom().String()
}

func (f *Factory) getKinds() map[reflect.Kind]FactoryFunc {
	f.kinds.init.Do(func() {
		f.kinds.mapping = make(map[reflect.Kind]FactoryFunc)
		f.kinds.mapping[reflect.Struct] = f.kindStruct
		f.kinds.mapping[reflect.Ptr] = f.kindPtr
		f.kinds.mapping[reflect.Map] = f.kindMap
		f.kinds.mapping[reflect.Slice] = f.kindSlice
		f.kinds.mapping[reflect.Array] = f.kindArray
		f.kinds.mapping[reflect.Chan] = f.kindChan
	})
	return f.kinds.mapping
}

func (f *Factory) kindStruct(T any) any {
	rStruct := reflect.New(reflect.TypeOf(T)).Elem()
	numField := rStruct.NumField()
	for i := 0; i < numField; i++ {
		field := rStruct.Field(i)
		structField := rStruct.Type().Field(i)

		if field.CanSet() && f.getConfig().CanPopulateStructField(structField) {
			if newValue := reflect.ValueOf(f.Create(field.Interface())); newValue.IsValid() {
				field.Set(newValue)
			}
		}
	}
	return rStruct.Interface()
}

func (f *Factory) kindPtr(T any) any {
	ptr := reflect.New(reflect.TypeOf(T).Elem())               // new ptr
	elemT := reflect.New(ptr.Type().Elem()).Elem().Interface() // new ptr value
	value := f.Create(elemT)
	ptr.Elem().Set(reflect.ValueOf(value)) // set ptr with a value
	return ptr.Interface()
}

func (f *Factory) kindMap(T any) any {
	rt := reflect.TypeOf(T)
	rv := reflect.MakeMap(rt)

	total := f.getRandom().IntN(7)
	for i := 0; i < total; i++ {
		key := f.Create(reflect.New(rt.Key()).Elem().Interface())
		value := f.Create(reflect.New(rt.Elem()).Elem().Interface())
		rv.SetMapIndex(reflect.ValueOf(key), reflect.ValueOf(value))
	}

	return rv.Interface()
}

func (f *Factory) kindSlice(T any) any {
	var (
		rtype  = reflect.TypeOf(T)
		rslice = reflect.MakeSlice(rtype, 0, 0)
		total  = f.getRandom().IntN(7)
		values []reflect.Value
	)
	for i := 0; i < total; i++ {
		v := f.Create(reflect.New(rtype.Elem()).Elem().Interface())
		values = append(values, reflect.ValueOf(v))
	}

	rslice = reflect.Append(rslice, values...)
	return rslice.Interface()
}

func (f *Factory) kindArray(T any) any {
	var (
		rtype  = reflect.TypeOf(T)
		rarray = reflect.New(rtype).Elem()
		total  = f.getRandom().IntN(rarray.Len())
	)
	for i := 0; i < total; i++ {
		v := f.Create(reflect.New(rtype.Elem()).Elem().Interface())
		rarray.Index(i).Set(reflect.ValueOf(v))
	}
	return rarray.Interface()
}

func (f *Factory) kindChan(T any) any {
	return reflect.MakeChan(reflect.TypeOf(T), 0).Interface()
}

func nextValue(value reflect.Value) reflect.Value {
	switch value.Type().Kind() {

	case reflect.Array:
		return reflect.New(value.Type()).Elem()

	case reflect.Slice:
		return reflect.MakeSlice(value.Type(), 0, 0)

	case reflect.Chan:
		return reflect.MakeChan(value.Type(), 0)

	case reflect.Map:
		return reflect.MakeMap(value.Type())

	case reflect.Ptr:
		return reflect.New(value.Type().Elem())

	case reflect.Uintptr:
		return reflect.ValueOf(uintptr(Random.Int()))

	default:
		//reflect.UnsafePointer
		//reflect.Interface
		//reflect.Func
		//
		// returns nil to avoid unsafe edge cases
		return reflect.ValueOf(nil)
	}
}

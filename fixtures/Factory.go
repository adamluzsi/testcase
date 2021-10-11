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
		mapping map[reflect.Type]factoryFunc
	}
	kinds struct {
		init    sync.Once
		mapping map[reflect.Kind]kindFunc
	}
	randomInit sync.Once
}

type (
	factoryFunc = func(context.Context) any
	kindFunc    = func(T interface{}, ctx context.Context) any
)

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

func (f *Factory) Fixture(T interface{}, ctx context.Context) (_T interface{}) {
	if T == nil {
		// type error panic will be solved after go generics support
		panic(`nil is not accepted as input[T] type`)
	}
	rt := reflect.TypeOf(T)
	typeFunc, ok := f.getTypes()[rt]
	if ok {
		return typeFunc(ctx)
	}
	if kindFunc, ok := f.getKinds()[rt.Kind()]; ok {
		return kindFunc(T, ctx)
	}
	return T
}

func (f *Factory) RegisterType(T any, ff factoryFunc) {
	f.getTypes()[reflect.TypeOf(T)] = ff
}

func (f *Factory) getTypes() map[reflect.Type]factoryFunc {
	f.types.init.Do(func() {
		f.types.mapping = make(map[reflect.Type]factoryFunc)
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

func (f *Factory) int(context.Context) any {
	return f.getRandom().Int()
}

func (f *Factory) int8(context.Context) any {
	return int8(f.getRandom().Int())
}

func (f *Factory) int16(context.Context) any {
	return int16(f.getRandom().Int())
}

func (f *Factory) int32(context.Context) any {
	return int32(f.getRandom().Int())
}

func (f *Factory) int64(context.Context) any {
	return int64(f.getRandom().Int())
}

func (f *Factory) uint(context.Context) any {
	return uint(f.getRandom().Int())
}

func (f *Factory) uint8(context.Context) any {
	return uint8(f.getRandom().Int())
}

func (f *Factory) uint16(context.Context) any {
	return uint16(f.getRandom().Int())
}

func (f *Factory) uint32(context.Context) any {
	return uint32(f.getRandom().Int())
}

func (f *Factory) uint64(context.Context) any {
	return uint64(f.getRandom().Int())
}

func (f *Factory) float32(context.Context) any {
	return f.getRandom().Float32()
}

func (f *Factory) float64(context.Context) any {
	return f.getRandom().Float64()
}

func (f *Factory) timeTime(context.Context) any {
	return f.getRandom().Time()
}

func (f *Factory) timeDuration(context.Context) any {
	return time.Duration(f.getRandom().IntBetween(int(time.Second), math.MaxInt32))
}

func (f *Factory) bool(context.Context) any {
	return f.getRandom().Bool()
}

func (f *Factory) string(context.Context) any {
	return f.getRandom().String()
}

func (f *Factory) getKinds() map[reflect.Kind]kindFunc {
	f.kinds.init.Do(func() {
		f.kinds.mapping = make(map[reflect.Kind]kindFunc)
		f.kinds.mapping[reflect.Struct] = f.kindStruct
		f.kinds.mapping[reflect.Ptr] = f.kindPtr
		f.kinds.mapping[reflect.Map] = f.kindMap
		f.kinds.mapping[reflect.Slice] = f.kindSlice
		f.kinds.mapping[reflect.Array] = f.kindArray
		f.kinds.mapping[reflect.Chan] = f.kindChan
	})
	return f.kinds.mapping
}

func (f *Factory) kindStruct(T any, ctx context.Context) any {
	rStruct := reflect.New(reflect.TypeOf(T)).Elem()
	numField := rStruct.NumField()
	for i := 0; i < numField; i++ {
		field := rStruct.Field(i)
		structField := rStruct.Type().Field(i)

		if field.CanSet() && f.getConfig().CanPopulateStructField(structField) {
			if newValue := reflect.ValueOf(f.Fixture(field.Interface(), ctx)); newValue.IsValid() {
				field.Set(newValue)
			}
		}
	}
	return rStruct.Interface()
}

func (f *Factory) kindPtr(T any, ctx context.Context) any {
	ptr := reflect.New(reflect.TypeOf(T).Elem())               // new ptr
	elemT := reflect.New(ptr.Type().Elem()).Elem().Interface() // new ptr value
	value := f.Fixture(elemT, ctx)
	ptr.Elem().Set(reflect.ValueOf(value)) // set ptr with a value
	return ptr.Interface()
}

func (f *Factory) kindMap(T any, ctx context.Context) any {
	rt := reflect.TypeOf(T)
	rv := reflect.MakeMap(rt)

	total := f.getRandom().IntN(7)
	for i := 0; i < total; i++ {
		key := f.Fixture(reflect.New(rt.Key()).Elem().Interface(), ctx)
		value := f.Fixture(reflect.New(rt.Elem()).Elem().Interface(), ctx)
		rv.SetMapIndex(reflect.ValueOf(key), reflect.ValueOf(value))
	}

	return rv.Interface()
}

func (f *Factory) kindSlice(T any, ctx context.Context) any {
	var (
		rtype  = reflect.TypeOf(T)
		rslice = reflect.MakeSlice(rtype, 0, 0)
		total  = f.getRandom().IntN(7)
		values []reflect.Value
	)
	for i := 0; i < total; i++ {
		v := f.Fixture(reflect.New(rtype.Elem()).Elem().Interface(), ctx)
		values = append(values, reflect.ValueOf(v))
	}

	rslice = reflect.Append(rslice, values...)
	return rslice.Interface()
}

func (f *Factory) kindArray(T any, ctx context.Context) any {
	var (
		rtype  = reflect.TypeOf(T)
		rarray = reflect.New(rtype).Elem()
		total  = f.getRandom().IntN(rarray.Len())
	)
	for i := 0; i < total; i++ {
		v := f.Fixture(reflect.New(rtype.Elem()).Elem().Interface(), ctx)
		rarray.Index(i).Set(reflect.ValueOf(v))
	}
	return rarray.Interface()
}

func (f *Factory) kindChan(T any, _ context.Context) any {
	return reflect.MakeChan(reflect.TypeOf(T), 0).Interface()
}

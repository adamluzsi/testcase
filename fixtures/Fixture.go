package fixtures

import (
	"context"
	"math"
	"reflect"
	"sync"
	"testing"
	"time"

	"github.com/adamluzsi/testcase/random"
)

type IFactory interface {
	Create(testing.TB, context.Context, any) any
	Context() context.Context
}

func NewFactory(rnd *random.Random) *Factory {
	return &Factory{Random: rnd}
}

type Factory struct {
	Random      *random.Random
	StubContext context.Context

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

type FactoryFunc func(testing.TB, context.Context, any) any

func (f *Factory) getRandom() *random.Random {
	f.randomInit.Do(func() {
		if f.Random == nil {
			f.Random = Random
		}
	})
	return f.Random
}

func (f *Factory) Create(tb testing.TB, ctx context.Context, T any) interface{} {
	if T == nil {
		tb.Fatal(`nil is not accepted as input[T] type`)
		return nil
	}
	rt := reflect.TypeOf(T)

	typeFunc, ok := f.getTypes()[rt]
	if ok {
		return typeFunc(tb, ctx, T)
	}

	if kindFunc, ok := f.getKinds()[rt.Kind()]; ok {
		return kindFunc(tb, ctx, T)
	}

	tb.Fatalf(`missing FactoryFunc for %T`, T)
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

func (f *Factory) int(tb testing.TB, ctx context.Context, T any) any {
	return f.getRandom().Int()
}

func (f *Factory) int8(tb testing.TB, ctx context.Context, T any) any {
	return int8(f.getRandom().Int())
}

func (f *Factory) int16(tb testing.TB, ctx context.Context, T any) any {
	return int16(f.getRandom().Int())
}

func (f *Factory) int32(tb testing.TB, ctx context.Context, T any) any {
	return int32(f.getRandom().Int())
}

func (f *Factory) int64(tb testing.TB, ctx context.Context, T any) any {
	return int64(f.getRandom().Int())
}

func (f *Factory) uint(tb testing.TB, ctx context.Context, T any) any {
	return uint(f.getRandom().Int())
}

func (f *Factory) uint8(tb testing.TB, ctx context.Context, T any) any {
	return uint8(f.getRandom().Int())
}

func (f *Factory) uint16(tb testing.TB, ctx context.Context, T any) any {
	return uint16(f.getRandom().Int())
}

func (f *Factory) uint32(tb testing.TB, ctx context.Context, T any) any {
	return uint32(f.getRandom().Int())
}

func (f *Factory) uint64(tb testing.TB, ctx context.Context, T any) any {
	return uint64(f.getRandom().Int())
}

func (f *Factory) float32(tb testing.TB, ctx context.Context, a any) any {
	return f.getRandom().Float32()
}

func (f *Factory) float64(tb testing.TB, ctx context.Context, a any) any {
	return f.getRandom().Float64()
}

func (f *Factory) timeTime(tb testing.TB, ctx context.Context, T any) any {
	return f.getRandom().Time()
}

func (f *Factory) timeDuration(tb testing.TB, ctx context.Context, T any) any {
	return time.Duration(f.getRandom().IntBetween(int(time.Second), math.MaxInt32))
}

func (f *Factory) bool(tb testing.TB, ctx context.Context, T any) any {
	return f.getRandom().Bool()
}

func (f *Factory) string(tb testing.TB, ctx context.Context, T any) any {
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

func (f *Factory) kindStruct(tb testing.TB, ctx context.Context, T any) any {
	rStruct := reflect.New(reflect.TypeOf(T)).Elem()
	numField := rStruct.NumField()
	for i := 0; i < numField; i++ {
		rField := rStruct.Field(i)
		v := f.Create(tb, ctx, rField.Interface())
		rField.Set(reflect.ValueOf(v))
	}
	return rStruct.Interface()
}

func (f *Factory) kindPtr(tb testing.TB, ctx context.Context, T any) any {
	ptr := reflect.New(reflect.TypeOf(T).Elem())               // new ptr
	elemT := reflect.New(ptr.Type().Elem()).Elem().Interface() // new ptr value
	value := f.Create(tb, ctx, elemT)
	ptr.Elem().Set(reflect.ValueOf(value)) // set ptr with a value
	return ptr.Interface()
}

func (f *Factory) kindMap(tb testing.TB, ctx context.Context, T any) any {
	rt := reflect.TypeOf(T)
	rv := reflect.MakeMap(rt)

	total := f.getRandom().IntN(7)
	for i := 0; i < total; i++ {
		key := f.Create(tb, ctx, reflect.New(rt.Key()).Elem().Interface())
		value := f.Create(tb, ctx, reflect.New(rt.Elem()).Elem().Interface())
		rv.SetMapIndex(reflect.ValueOf(key), reflect.ValueOf(value))
	}

	return rv.Interface()
}

func (f *Factory) kindSlice(tb testing.TB, ctx context.Context, T any) any {
	var (
		rtype  = reflect.TypeOf(T)
		rslice = reflect.MakeSlice(rtype, 0, 0)
		total  = f.getRandom().IntN(7)
		values []reflect.Value
	)
	for i := 0; i < total; i++ {
		v := f.Create(tb, ctx, reflect.New(rtype.Elem()).Elem().Interface())
		values = append(values, reflect.ValueOf(v))
	}

	rslice = reflect.Append(rslice, values...)
	return rslice.Interface()
}

func (f *Factory) kindArray(tb testing.TB, ctx context.Context, T any) any {
	var (
		rtype  = reflect.TypeOf(T)
		rarray = reflect.New(rtype).Elem()
		total  = f.getRandom().IntN(rarray.Len())
	)
	for i := 0; i < total; i++ {
		v := f.Create(tb, ctx, reflect.New(rtype.Elem()).Elem().Interface())
		rarray.Index(i).Set(reflect.ValueOf(v))
	}
	return rarray.Interface()
}

func (f *Factory) kindChan(tb testing.TB, ctx context.Context, T any) any {
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

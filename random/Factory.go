package random

import (
	"math"
	"reflect"
	"sync"
	"time"
)

type Factory struct {
	types struct {
		init    sync.Once
		mapping map[reflect.Type]typeFunc
	}
	kinds struct {
		init    sync.Once
		mapping map[reflect.Kind]kindFunc
	}
}

type (
	typeFunc func(r *Random) any
	kindFunc func(r *Random, T interface{}) any
)

func (f *Factory) Make(rnd *Random, T any) (_T any) {
	if T == nil {
		panic(`nil is not accepted value type`)
	}
	rt := reflect.TypeOf(T)
	if typeFunc, ok := f.getTypes()[rt]; ok {
		return typeFunc(rnd)
	}
	if kindFunc, ok := f.getKinds()[rt.Kind()]; ok {
		return kindFunc(rnd, T)
	}
	return T
}

func (f *Factory) RegisterType(T any, ff typeFunc) {
	f.getTypes()[reflect.TypeOf(T)] = ff
}

func (f *Factory) getTypes() map[reflect.Type]typeFunc {
	f.types.init.Do(func() {
		f.types.mapping = make(map[reflect.Type]typeFunc)
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
		f.types.mapping[reflect.TypeOf(uintptr(0))] = f.uintptr
		f.types.mapping[reflect.TypeOf(string(""))] = f.string
		f.types.mapping[reflect.TypeOf(bool(false))] = f.bool
		f.types.mapping[reflect.TypeOf(time.Time{})] = f.timeTime
		f.types.mapping[reflect.TypeOf(time.Duration(0))] = f.timeDuration
	})
	return f.types.mapping
}

func (f *Factory) getRandom() *Random {
	return New(CryptoSeed{})
}

func (f *Factory) int(rnd *Random) any {
	return rnd.Int()
}

func (f *Factory) int8(rnd *Random) any {
	return int8(rnd.Int())
}

func (f *Factory) int16(rnd *Random) any {
	return int16(rnd.Int())
}

func (f *Factory) int32(rnd *Random) any {
	return int32(rnd.Int())
}

func (f *Factory) int64(rnd *Random) any {
	return int64(rnd.Int())
}

func (f *Factory) uint(rnd *Random) any {
	return uint(rnd.Int())
}

func (f *Factory) uint8(rnd *Random) any {
	return uint8(rnd.Int())
}

func (f *Factory) uint16(rnd *Random) any {
	return uint16(rnd.Int())
}

func (f *Factory) uint32(rnd *Random) any {
	return uint32(rnd.Int())
}

func (f *Factory) uint64(rnd *Random) any {
	return uint64(rnd.Int())
}

func (f *Factory) float32(rnd *Random) any {
	return rnd.Float32()
}

func (f *Factory) float64(rnd *Random) any {
	return rnd.Float64()
}

func (f *Factory) uintptr(rnd *Random) any {
	return uintptr(rnd.Int())
}

func (f *Factory) timeTime(rnd *Random) any {
	return rnd.Time()
}

func (f *Factory) timeDuration(rnd *Random) any {
	return time.Duration(rnd.IntBetween(int(time.Second), math.MaxInt32))
}

func (f *Factory) bool(rnd *Random) any {
	return rnd.Bool()
}

func (f *Factory) string(rnd *Random) any {
	return rnd.String()
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

func (f *Factory) kindStruct(rnd *Random, T any) any {
	rStruct := reflect.New(reflect.TypeOf(T)).Elem()
	numField := rStruct.NumField()
	for i := 0; i < numField; i++ {
		field := rStruct.Field(i)
		// structField := rStruct.Type().Field(i)

		if field.CanSet() /* && f.getConfig().CanPopulateStructField(structField) */ {
			if newValue := reflect.ValueOf(f.Make(rnd, field.Interface())); newValue.IsValid() {
				field.Set(newValue)
			}
		}
	}
	return rStruct.Interface()
}

func (f *Factory) kindPtr(rnd *Random, T any) any {
	ptr := reflect.New(reflect.TypeOf(T).Elem())               // new ptr
	elemT := reflect.New(ptr.Type().Elem()).Elem().Interface() // new ptr value
	value := f.Make(rnd, elemT)
	ptr.Elem().Set(reflect.ValueOf(value)) // set ptr with a value
	return ptr.Interface()
}

func (f *Factory) kindMap(rnd *Random, T any) any {
	rt := reflect.TypeOf(T)
	rv := reflect.MakeMap(rt)

	total := rnd.IntN(7)
	for i := 0; i < total; i++ {
		key := f.Make(rnd, reflect.New(rt.Key()).Elem().Interface())
		value := f.Make(rnd, reflect.New(rt.Elem()).Elem().Interface())
		rv.SetMapIndex(reflect.ValueOf(key), reflect.ValueOf(value))
	}

	return rv.Interface()
}

func (f *Factory) kindSlice(rnd *Random, T any) any {
	var (
		rtype  = reflect.TypeOf(T)
		rslice = reflect.MakeSlice(rtype, 0, 0)
		total  = rnd.IntN(7)
		values []reflect.Value
	)
	for i := 0; i < total; i++ {
		v := f.Make(rnd, reflect.New(rtype.Elem()).Elem().Interface())
		values = append(values, reflect.ValueOf(v))
	}

	rslice = reflect.Append(rslice, values...)
	return rslice.Interface()
}

func (f *Factory) kindArray(rnd *Random, T any) any {
	var (
		rtype  = reflect.TypeOf(T)
		rarray = reflect.New(rtype).Elem()
		total  = rnd.IntN(rarray.Len())
	)
	for i := 0; i < total; i++ {
		v := f.Make(rnd, reflect.New(rtype.Elem()).Elem().Interface())
		rarray.Index(i).Set(reflect.ValueOf(v))
	}
	return rarray.Interface()
}

func (f *Factory) kindChan(rnd *Random, T any) any {
	return reflect.MakeChan(reflect.TypeOf(T), 0).Interface()
}

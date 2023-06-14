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
	kindFunc func(r *Random, T reflect.Type) any
)

func (f *Factory) Make(rnd *Random, T any) (_T any) {
	if T == nil {
		return nil
	}
	rt := reflect.TypeOf(T)
	if typeFunc, ok := f.getTypes()[rt]; ok {
		return typeFunc(rnd)
	}
	if kindFunc, ok := f.getKinds()[rt.Kind()]; ok {
		return kindFunc(rnd, reflect.TypeOf(T))
	}
	return T
}

func (f *Factory) RegisterType(T any, ff typeFunc) {
	f.getTypes()[reflect.TypeOf(T)] = ff
}

func (f *Factory) getTypes() map[reflect.Type]typeFunc {
	f.types.init.Do(func() {
		f.types.mapping = make(map[reflect.Type]typeFunc)
		f.types.mapping[reflect.TypeOf(time.Time{})] = f.timeTime
		f.types.mapping[reflect.TypeOf(time.Duration(0))] = f.timeDuration
	})
	return f.types.mapping
}

func (f *Factory) getRandom() *Random {
	return New(CryptoSeed{})
}

func (f *Factory) int(rnd *Random, T reflect.Type) any {
	return reflect.ValueOf(rnd.Int()).Convert(T).Interface()
}

func (f *Factory) int8(rnd *Random, T reflect.Type) any {
	return reflect.ValueOf(int8(rnd.Int())).Convert(T).Interface()
}

func (f *Factory) int16(rnd *Random, T reflect.Type) any {
	return reflect.ValueOf(int16(rnd.Int())).Convert(T).Interface()
}

func (f *Factory) int32(rnd *Random, T reflect.Type) any {
	return reflect.ValueOf(int32(rnd.Int())).Convert(T).Interface()
}

func (f *Factory) int64(rnd *Random, T reflect.Type) any {
	return reflect.ValueOf(int64(rnd.Int())).Convert(T).Interface()
}

func (f *Factory) uint(rnd *Random, T reflect.Type) any {
	return reflect.ValueOf(uint(rnd.Int())).Convert(T).Interface()
}

func (f *Factory) uint8(rnd *Random, T reflect.Type) any {
	return reflect.ValueOf(uint8(rnd.Int())).Convert(T).Interface()
}

func (f *Factory) uint16(rnd *Random, T reflect.Type) any {
	return reflect.ValueOf(uint16(rnd.Int())).Convert(T).Interface()
}

func (f *Factory) uint32(rnd *Random, T reflect.Type) any {
	return reflect.ValueOf(uint32(rnd.Int())).Convert(T).Interface()
}

func (f *Factory) uint64(rnd *Random, T reflect.Type) any {
	return reflect.ValueOf(uint64(rnd.Int())).Convert(T).Interface()
}

func (f *Factory) float32(rnd *Random, T reflect.Type) any {
	return reflect.ValueOf(rnd.Float32()).Convert(T).Interface()
}

func (f *Factory) float64(rnd *Random, T reflect.Type) any {
	return reflect.ValueOf(rnd.Float64()).Convert(T).Interface()
}

func (f *Factory) uintptr(rnd *Random, T reflect.Type) any {
	return reflect.ValueOf(uintptr(rnd.Int())).Convert(T).Interface()
}

func (f *Factory) timeTime(rnd *Random) any {
	return rnd.Time()
}

func (f *Factory) timeDuration(rnd *Random) any {
	return time.Duration(rnd.IntBetween(int(time.Second), math.MaxInt32))
}

func (f *Factory) bool(rnd *Random, T reflect.Type) any {
	return reflect.ValueOf(rnd.Bool()).Convert(T).Interface()
}

func (f *Factory) string(rnd *Random, T reflect.Type) any {
	return reflect.ValueOf(rnd.String()).Convert(T).Interface()
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
		f.kinds.mapping[reflect.Int] = f.int
		f.kinds.mapping[reflect.Int8] = f.int8
		f.kinds.mapping[reflect.Int16] = f.int16
		f.kinds.mapping[reflect.Int32] = f.int32
		f.kinds.mapping[reflect.Int64] = f.int64
		f.kinds.mapping[reflect.Uint] = f.uint
		f.kinds.mapping[reflect.Uint8] = f.uint8
		f.kinds.mapping[reflect.Uint16] = f.uint16
		f.kinds.mapping[reflect.Uint32] = f.uint32
		f.kinds.mapping[reflect.Uint64] = f.uint64
		f.kinds.mapping[reflect.Float32] = f.float32
		f.kinds.mapping[reflect.Float64] = f.float64
		f.kinds.mapping[reflect.Uintptr] = f.uintptr
		f.kinds.mapping[reflect.String] = f.string
		f.kinds.mapping[reflect.Bool] = f.bool
	})
	return f.kinds.mapping
}

func (f *Factory) kindStruct(rnd *Random, T reflect.Type) any {
	rStruct := reflect.New(T).Elem()
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

func (f *Factory) kindPtr(rnd *Random, T reflect.Type) any {
	ptr := reflect.New(T.Elem())                               // new ptr
	elemT := reflect.New(ptr.Type().Elem()).Elem().Interface() // new ptr value
	value := f.Make(rnd, elemT)
	ptr.Elem().Set(reflect.ValueOf(value)) // set ptr with a value
	return ptr.Interface()
}

func (f *Factory) kindMap(rnd *Random, T reflect.Type) any {
	rv := reflect.MakeMap(T)

	total := rnd.IntN(7)
	for i := 0; i < total; i++ {
		key := f.Make(rnd, reflect.New(T.Key()).Elem().Interface())
		value := f.Make(rnd, reflect.New(T.Elem()).Elem().Interface())
		rv.SetMapIndex(reflect.ValueOf(key), reflect.ValueOf(value))
	}

	return rv.Interface()
}

func (f *Factory) kindSlice(rnd *Random, T reflect.Type) any {
	var (
		rslice = reflect.MakeSlice(T, 0, 0)
		total  = rnd.IntN(7)
		values []reflect.Value
	)
	for i := 0; i < total; i++ {
		v := f.Make(rnd, reflect.New(T.Elem()).Elem().Interface())
		values = append(values, reflect.ValueOf(v))
	}

	rslice = reflect.Append(rslice, values...)
	return rslice.Interface()
}

func (f *Factory) kindArray(rnd *Random, T reflect.Type) any {
	var (
		rarray = reflect.New(T).Elem()
		total  = rnd.IntN(rarray.Len())
	)
	for i := 0; i < total; i++ {
		v := f.Make(rnd, reflect.New(T.Elem()).Elem().Interface())
		rarray.Index(i).Set(reflect.ValueOf(v))
	}
	return rarray.Interface()
}

func (f *Factory) kindChan(rnd *Random, T reflect.Type) any {
	return reflect.MakeChan(T, 0).Interface()
}

package random

import (
	"math/rand"
	"reflect"
	"sync"
	"time"
)

func NewRandom(s rand.Source) *Random {
	return &Random{Source: s}
}

// A Random is a source of random numbers.
// It is safe to be used in from multiple goroutines.
type Random struct {
	Source rand.Source

	m sync.Mutex
}

// Int returns a non-negative pseudo-random int.
func (r *Random) Int() int {
	r.m.Lock()
	defer r.m.Unlock()
	return rand.New(r.Source).Int()
}

// IntN returns, as an int, a non-negative pseudo-random number in [0,n).
// It panics if n <= 0.
func (r *Random) IntN(n int) int {
	r.m.Lock()
	defer r.m.Unlock()
	return rand.New(r.Source).Intn(n)
}

// Float32 returns, as a float32, a pseudo-random number in [0.0,1.0).
func (r *Random) Float32() float32 {
	r.m.Lock()
	defer r.m.Unlock()
	return rand.New(r.Source).Float32()
}

// Float64 returns, as a float64, a pseudo-random number in [0.0,1.0).
func (r *Random) Float64() float64 {
	r.m.Lock()
	defer r.m.Unlock()
	return rand.New(r.Source).Float64()
}

// IntBetween returns, as an int, a non-negative pseudo-random number based on the received int range's [min,max].
func (r *Random) IntBetween(min, max int) int {
	return min + r.IntN((max+1)-min)
}

func (r *Random) ElementFromSlice(slice interface{}) interface{} {
	s := reflect.ValueOf(slice)
	index := rand.New(r.Source).Intn(s.Len())
	return s.Index(index).Interface()
}

func (r *Random) KeyFromMap(anyMap interface{}) interface{} {
	s := reflect.ValueOf(anyMap)
	index := rand.New(r.Source).Intn(s.Len())
	return s.MapKeys()[index].Interface()
}

func (r *Random) Bool() bool {
	return r.IntN(2) == 0
}

func (r *Random) String() string {
	return r.StringN(r.IntBetween(4, 42))
}

func (r *Random) StringN(length int) string {
	r.m.Lock()
	defer r.m.Unlock()

	const letters = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz-"

	bytes := make([]byte, length)
	if _, err := rand.New(r.Source).Read(bytes); err != nil {
		panic(err)
	}

	for i, b := range bytes {
		bytes[i] = letters[b%byte(len(letters))]
	}

	return string(bytes)
}

// TimeBetween returns, as an time.Time, a non-negative pseudo-random time in [from,to].
func (r *Random) TimeBetween(from, to time.Time) time.Time {
	return time.Unix(int64(r.IntBetween(int(from.Unix()), int(to.Unix()))), 0).UTC()
}

func (r *Random) Time() time.Time {
	t := time.Now().UTC()
	from := t.AddDate(0, 0, r.IntN(42)*-1)
	to := t.AddDate(0, 0, r.IntN(42)).Add(time.Second)
	return r.TimeBetween(from, to)
}

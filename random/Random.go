package random

import (
	"errors"
	"fmt"
	"github.com/adamluzsi/testcase/random/internal"
	"math/rand"
	"reflect"
	"strconv"
	"strings"
	"sync"
	"time"
)

func New(s rand.Source) *Random {
	return &Random{Source: s}
}

// A Random is a source of random numbers.
// It is safe to be used in from multiple goroutines.
type Random struct {
	Source  rand.Source
	Factory Factory

	m sync.Mutex
}

func (r *Random) rnd() *rand.Rand {
	return rand.New(r.Source)
}

// Int returns a non-negative pseudo-random int.
func (r *Random) Int() int {
	r.m.Lock()
	defer r.m.Unlock()
	return r.rnd().Int()
}

// IntN returns, as an int, a non-negative pseudo-random number in [0,n).
// It panics if n <= 0.
func (r *Random) IntN(n int) int {
	r.m.Lock()
	defer r.m.Unlock()
	return r.rnd().Intn(n)
}

// Float32 returns, as a float32, a pseudo-random number in [0.0,1.0).
func (r *Random) Float32() float32 {
	r.m.Lock()
	defer r.m.Unlock()
	return r.rnd().Float32()
}

// Float64 returns, as a float64, a pseudo-random number in [0.0,1.0).
func (r *Random) Float64() float64 {
	r.m.Lock()
	defer r.m.Unlock()
	return r.rnd().Float64()
}

// IntBetween returns, as an int, a non-negative pseudo-random number based on the received int range's [min,max].
func (r *Random) IntBetween(min, max int) int {
	return min + r.IntN((max+1)-min)
}

// IntB returns, as an int, a non-negative pseudo-random number based on the received int range's [min,max].
func (r *Random) IntB(min, max int) int {
	return r.IntBetween(min, max)
}

func (r *Random) ElementFromSlice(slice interface{}) interface{} {
	s := reflect.ValueOf(slice)
	index := r.rnd().Intn(s.Len())
	return s.Index(index).Interface()
}

func (r *Random) KeyFromMap(anyMap interface{}) interface{} {
	s := reflect.ValueOf(anyMap)
	index := r.rnd().Intn(s.Len())
	return s.MapKeys()[index].Interface()
}

func (r *Random) Bool() bool {
	return r.IntN(2) == 0
}

func (r *Random) Error() error {
	msg := fixtureStrings.errors[r.IntN(len(fixtureStrings.errors))]
	return errors.New(msg)
}

func (r *Random) String() string {
	return fixtureStrings.naughty[r.IntN(len(fixtureStrings.naughty))]
}

func (r *Random) StringN(length int) string {
	return r.StringNWithCharset(length, charset)
}

func (r *Random) StringNC(length int, charset string) string {
	return r.StringNWithCharset(length, charset)
}

func (r *Random) StringNWithCharset(length int, charset string) string {
	r.m.Lock()
	defer r.m.Unlock()

	bytes := make([]byte, length)

	if _, err := r.rnd().Read(bytes); err != nil {
		panic(err)
	}

	for i, b := range bytes {
		bytes[i] = charset[b%byte(len(charset))]
	}

	return string(bytes)
}

// TimeBetween returns, as an time.Time, a non-negative pseudo-random time in [from,to].
func (r *Random) TimeBetween(from, to time.Time) time.Time {
	return time.Unix(int64(r.IntBetween(int(from.Unix()), int(to.Unix()))), 0).UTC()
}

// TimeB returns, as an time.Time, a non-negative pseudo-random time in [from,to].
func (r *Random) TimeB(from, to time.Time) time.Time {
	return r.TimeBetween(from, to)
}

func (r *Random) Time() time.Time {
	t := time.Now().UTC()
	from := t.AddDate(0, 0, r.IntN(42)*-1)
	to := t.AddDate(0, 0, r.IntN(42)).Add(time.Second)
	return r.TimeBetween(from, to)
}

func (r *Random) TimeN(from time.Time, years, months, days int) time.Time {
	nIntN := func(n int) int {
		if n == 0 {
			return 0
		}
		if n < 0 {
			return r.IntN(n*-1) * -1
		}
		return r.IntN(n)
	}

	base := time.Date(from.Year(), from.Month(), from.Day(), from.Hour(), from.Minute(), from.Second(), 0, from.Location())
	return base.AddDate(nIntN(years), nIntN(months), nIntN(days))
}

func (r *Random) Read(p []byte) (n int, err error) {
	return r.rnd().Read(p)
}

func (r *Random) mustRead(b []byte) {
	var err error
	for i := 0; i < 42; i++ {
		_, err = r.Read(b)
		if err == nil {
			return
		}
	}
	panic(err)
}

func (r *Random) UUID() string {
	b := make([]byte, 16)
	r.mustRead(b)
	return fmt.Sprintf("%X-%X-%X-%X-%X", b[0:4], b[4:6], b[6:8], b[8:10], b[10:])
}

func (r *Random) Name() name {
	return name{Random: r}
}

type name struct{ Random *Random }

func (n name) First(opts ...internal.PersonOption) string {
	sexType := internal.ToPersonConfig(opts...).SexType
	switch sexType {
	case internal.SexTypeAny, 0:
		sexType = randomSexType(n.Random)
	}
	switch sexType {
	case internal.SexTypeMale:
		return n.Random.ElementFromSlice(fixtureStrings.names.male).(string)
	case internal.SexTypeFemale:
		return n.Random.ElementFromSlice(fixtureStrings.names.female).(string)
	default:
		panic("not implemented")
	}
}

func (n name) Last() string {
	return n.Random.ElementFromSlice(fixtureStrings.names.last).(string)
}

func (r *Random) Email() string {
	return fmt.Sprintf("%s%s%s@%s",
		strings.ToLower(r.Name().First()),
		strings.ToLower(r.Name().Last()),
		strconv.Itoa(r.IntB(0, 32)),
		r.ElementFromSlice(fixtureStrings.emailDomains).(string),
	)
}

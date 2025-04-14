package random

import (
	"errors"
	"fmt"
	"math"
	"math/big"
	"math/rand"
	"reflect"
	"strconv"
	"strings"
	"sync"
	"time"

	"go.llib.dev/testcase/internal"
)

var defaultRandom = New(CryptoSeed{})

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

func (r *Random) FloatN(n float64) float64 {
	if n <= 0 {
		panic("invalid argument to FloatN")
	}
	return r.Float64() * n
}

// FloatBetween returns a float between the given min and max value range. [min, max]
func (r *Random) FloatBetween(min, max float64) float64 {
	const (
		whenMin = 0
		whenMax = 1000
	)
	switch r.IntB(whenMin, whenMax) {
	case whenMin:
		return min
	case whenMax:
		return max
	default: // when between min and max
		return min + r.Float64()*(max-min)
	}
}

// FloatB returns a float between the given min and max value range.
func (r *Random) FloatB(min, max float64) float64 {
	return r.FloatBetween(min, max)
}

// IntBetween returns an int based on the received int range's [min,max].
func (r *Random) IntBetween(min, max int) int {
	min, max = correct(min, max)
	// check if intn(max+1) would overflow.
	// we need this in order to make max part of the valid result set.
	if canOverflow(max, 1) {
		// shift the valid result set to left by 1
		//   min - 1, math.MaxInt - 1
		// then shift the reult back to the original position.
		return r.IntBetween(min-1, max-1) + 1
	}
	// check if max-min would overflow
	// we need this in order to convert intn into intb
	// by adding the min to the result of intn(max-min+1)
	if canOverflow(-min, max) {
		return r.overflowIntBetween(min, max)
	}

	return r.intb(min, max)
}

func (r *Random) intb(min int, max int) int {
	min, max = correct(min, max)
	return min + r.IntN(max+1-min)
}

func (r *Random) overflowIntBetween(min int, max int) int {
	min, max = correct(min, max)
	a := big.NewInt(int64(min))
	b := big.NewInt(int64(max))
	middle := new(big.Int).Sub(b, a)
	middle.Div(middle, big.NewInt(2))
	boundary := int(new(big.Int).Add(a, middle).Int64())
	if r.Bool() {
		return r.intb(min, boundary)
	}
	return r.intb(boundary, max)
}

func canOverflow(less, more int) bool {
	if more < less {
		less, more = more, less
	}
	switch {
	case 0 < less && 0 < more:
		// MinInt - -number -> MinInt plus abs less
		maxLess := math.MaxInt - more
		return maxLess < less // positive overflow
	case less < 0 && more < 0:
		// MinInt - -number -> MinInt plus abs less
		minMore := math.MinInt - less
		return more < minMore // negative overflow
	case less < 0 && 0 < more:
		// there is no combination where a + b can cause overflow
		// because even MinInt plus MaxInt would only end up in zero.
	}
	return false //, 0, more + minMore
}

// DurationB returns an duration based on the received duration range's [min,max].
func (r *Random) DurationB(min, max time.Duration) time.Duration {
	return r.DurationBetween(min, max)
}

// DurationBetween returns an duration based on the received duration range's [min,max].
func (r *Random) DurationBetween(min, max time.Duration) time.Duration {
	return time.Duration(r.IntBetween(int(min), int(max)))
}

// IntB returns, as an int, a non-negative pseudo-random number based on the received int range's [min,max].
func (r *Random) IntB(min, max int) int {
	return r.IntBetween(min, max)
}

// Pick will return a random element picked from a slice.
// You need type assert the returned value to get back the original type.
func (r *Random) Pick(slice any) any {
	s := reflect.ValueOf(slice)
	index := r.rnd().Intn(s.Len())
	return s.Index(index).Interface()
}

func Pick[T any](rnd *Random, vs ...T) T {
	if rnd == nil {
		rnd = defaultRandom
	}
	return rnd.Pick(vs).(T)
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
	if to.Before(from) {
		from, to = to, from
	}
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
	deadline := time.Now().Add(5 * time.Minute)
	for {
		n, err := r.Read(b)
		if err != nil {
			if time.Now().After(deadline) {
				panic(err)
			}
			continue
		}
		if n == len(b) {
			return
		}
	}
}

func (r *Random) UUID() string {
	b := make([]byte, 16)
	r.mustRead(b)
	return fmt.Sprintf("%X-%X-%X-%X-%X", b[0:4], b[4:6], b[6:8], b[8:10], b[10:])
}

type Contact struct {
	FirstName string
	LastName  string
	Email     string
}

func (r *Random) Contact(opts ...internal.ContactOption) Contact {
	conf := internal.ToContactConfig(opts...)
	cg := contactGenerator{Random: r}
	var c Contact
	c.FirstName = cg.first(conf)
	c.LastName = cg.last()
	c.Email = cg.email(c.FirstName, c.LastName)
	return c
}

type contactGenerator struct{ Random *Random }

func (cg contactGenerator) first(conf internal.ContactConfig) string {
	sexType := conf.SexType
	switch sexType {
	case internal.SexTypeAny, 0:
		sexType = randomSexType(cg.Random)
	}
	switch sexType {
	case internal.SexTypeMale:
		return cg.Random.Pick(fixtureStrings.names.male).(string)
	case internal.SexTypeFemale:
		return cg.Random.Pick(fixtureStrings.names.female).(string)
	default:
		panic("not implemented")
	}
}

func (cg contactGenerator) last() string {
	return cg.Random.Pick(fixtureStrings.names.last).(string)
}

func (cg contactGenerator) email(firstName, lastName string) string {
	return fmt.Sprintf("%s%s%s%s@%s",
		strings.ToLower(firstName),
		cg.Random.Pick([]string{"_", "."}).(string),
		strings.ToLower(lastName),
		strconv.Itoa(cg.Random.IntB(0, 42)),
		cg.Random.Pick(fixtureStrings.emailDomains).(string))
}

// Repeat will repeatedly call the "do" function.
// The number of repeats will be random between the min and the max range.
// The repeated time will be returned as a result.
func (r *Random) Repeat(min, max int, do func()) int {
	n := r.IntB(min, max)
	for i := 0; i < n; i++ {
		do()
	}
	return n
}

// Domain will return a valid domain name.
func (r *Random) Domain() string {
	return r.Pick(fixtureStrings.domains).(string)
}

type number interface {
	int
}

func correct[N number](min, max N) (N, N) {
	if max < min {
		return max, min
	}
	return min, max
}

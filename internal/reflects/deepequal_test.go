package reflects_test

import (
	"errors"
	"math"
	"reflect"
	"testing"
	"time"

	"go.llib.dev/testcase/internal/reflects"
	"go.llib.dev/testcase/random"
)

func TestDeepEqual(t *testing.T) {
	rnd := random.New(random.CryptoSeed{})
	expErr := rnd.Error()
	var (
		sharedPtr = intPtr(42)
		sameChan  = bufferedChan(1, 2)
		backing1  = []int{1, 2}
		backing2  = []int{1, 3}
	)
	tt := []struct {
		desc     string
		v1, v2   any
		isEqual  bool
		hasError error
	}{
		{
			desc:    "equal integers",
			v1:      1,
			v2:      1,
			isEqual: true,
		},
		{
			desc:    "different integers",
			v1:      1,
			v2:      2,
			isEqual: false,
		},
		{
			desc:    "equal strings",
			v1:      "test",
			v2:      "test",
			isEqual: true,
		},
		{
			desc:    "different strings",
			v1:      "test",
			v2:      "test1",
			isEqual: false,
		},
		{
			desc:    "equal slices",
			v1:      []int{1, 2, 3},
			v2:      []int{1, 2, 3},
			isEqual: true,
		},
		{
			desc:    "different slices",
			v1:      []int{1, 2, 3},
			v2:      []int{1, 2, 4},
			isEqual: false,
		},
		{
			desc:    "equal arrays",
			v1:      [3]int{1, 2, 3},
			v2:      [3]int{1, 2, 3},
			isEqual: true,
		},
		{
			desc:    "different arrays",
			v1:      [3]int{1, 2, 3},
			v2:      [3]int{1, 2, 4},
			isEqual: false,
		},
		{
			desc:    "equal maps",
			v1:      map[string]int{"one": 1, "two": 2},
			v2:      map[string]int{"one": 1, "two": 2},
			isEqual: true,
		},
		{
			desc:    "different maps",
			v1:      map[string]int{"one": 1, "two": 2},
			v2:      map[string]int{"one": 1, "two": 3},
			isEqual: false,
		},
		{
			desc:    "equal structs",
			v1:      Struct{Field1: 1, Field2: "test"},
			v2:      Struct{Field1: 1, Field2: "test"},
			isEqual: true,
		},
		{
			desc:    "different structs",
			v1:      Struct{Field1: 1, Field2: "test"},
			v2:      Struct{Field1: 2, Field2: "test"},
			isEqual: false,
		},
		{
			desc: "different structs with equality support - equal",
			v1: StructWithMethodEqual{
				Irrelevant: rnd.Int(),
				Relevant:   42,
			},
			v2: StructWithMethodEqual{
				Irrelevant: rnd.Int(),
				Relevant:   42,
			},
			isEqual: true,
		},
		{
			desc: "different structs with equality support - not equal",
			v1: StructWithMethodEqual{
				Irrelevant: rnd.Int(),
				Relevant:   24,
			},
			v2: StructWithMethodEqual{
				Irrelevant: rnd.Int(),
				Relevant:   42,
			},
			isEqual: false,
		},
		{
			desc: "different structs with equality support (ptr receiver) - equal",
			v1: TestStructEquatableOnPtr{
				Irrelevant: rnd.Int(),
				Relevant:   42,
			},
			v2: TestStructEquatableOnPtr{
				Irrelevant: rnd.Int(),
				Relevant:   42,
			},
			isEqual: true,
		},
		{
			desc: "different structs with equality support (ptr receiver) - not equal",
			v1: TestStructEquatableOnPtr{
				Irrelevant: rnd.Int(),
				Relevant:   24,
			},
			v2: TestStructEquatableOnPtr{
				Irrelevant: rnd.Int(),
				Relevant:   42,
			},
			isEqual: false,
		},
		{
			desc: "different structs with equality support (IsEqual) - equal",
			v1: StructWithMethodIsEqual{
				Irrelevant: rnd.Int(),
				Relevant:   42,
			},
			v2: StructWithMethodIsEqual{
				Irrelevant: rnd.Int(),
				Relevant:   42,
			},
			isEqual: true,
		},
		{
			desc: "different structs with equality support (IsEqual) - not equal",
			v1: StructWithMethodIsEqual{
				Irrelevant: rnd.Int(),
				Relevant:   24,
			},
			v2: StructWithMethodIsEqual{
				Irrelevant: rnd.Int(),
				Relevant:   42,
			},
			isEqual: false,
		},
		{
			desc: "different structs with equality+err support - equal",
			v1: StructWithMethodEqualWithErr{
				Irrelevant: rnd.Int(),
				Relevant:   42,
				EqualErr:   nil,
			},
			v2: StructWithMethodEqualWithErr{
				Irrelevant: rnd.Int(),
				Relevant:   42,
				EqualErr:   nil,
			},
			isEqual: true,
		},
		{
			desc: "different structs with equality+err support - has error",
			v1: StructWithMethodEqualWithErr{
				Irrelevant: rnd.Int(),
				Relevant:   42,
				EqualErr:   expErr,
			},
			v2: StructWithMethodEqualWithErr{
				Irrelevant: rnd.Int(),
				Relevant:   42,
				EqualErr:   expErr,
			},
			isEqual:  false,
			hasError: expErr,
		},
		{
			desc: "different structs with comparable support - equal",
			v1: StructWithMethodCmp{
				Irrelevant: rnd.Int(),
				Relevant:   42,
			},
			v2: StructWithMethodCmp{
				Irrelevant: rnd.Int(),
				Relevant:   42,
			},
			isEqual: true,
		},
		{
			desc: "different structs with comparable support - not equal",
			v1: StructWithMethodCmp{
				Irrelevant: rnd.Int(),
				Relevant:   24,
			},
			v2: StructWithMethodCmp{
				Irrelevant: rnd.Int(),
				Relevant:   42,
			},
			isEqual: false,
		},
		{
			desc: "different structs with comparable support (ptr receiver) - equal",
			v1: StructWithCmpOnPtr{
				Irrelevant: rnd.Int(),
				Relevant:   42,
			},
			v2: StructWithCmpOnPtr{
				Irrelevant: rnd.Int(),
				Relevant:   42,
			},
			isEqual: true,
		},
		{
			desc: "different structs with comparable support (ptr receiver) - not equal",
			v1: StructWithPtrMethodCmp{
				Irrelevant: rnd.Int(),
				Relevant:   24,
			},
			v2: StructWithPtrMethodCmp{
				Irrelevant: rnd.Int(),
				Relevant:   42,
			},
			isEqual: false,
		},
		{
			desc: "structs - comparable - equal",
			v1: ComparableStruct{
				V: "foo",
			},
			v2: ComparableStruct{
				V: "foo",
			},
			isEqual: true,
		},
		{
			desc: "structs - comparable - not equal",
			v1: ComparableStruct{
				V: "foo",
			},
			v2: ComparableStruct{
				V: "bar",
			},
			isEqual: false,
		},
		{
			desc: "structs - comparable - not equal unexported",
			v1: ComparableStruct{
				V: "foo",
				v: "bar",
			},
			v2: ComparableStruct{
				V: "foo",
				v: "baz",
			},
			isEqual: false,
		},

		{
			desc: "structs - not comparable - equal",
			v1: NotComparableStruct{
				V: "foo",
			},
			v2: NotComparableStruct{
				V: "foo",
			},
			isEqual: true,
		},
		{
			desc: "structs - not comparable - not equal",
			v1: NotComparableStruct{
				V: "foo",
			},
			v2: NotComparableStruct{
				V: "bar",
			},
			isEqual: false,
		},
		{
			desc: "structs - not comparable - not equal unexported",
			v1: NotComparableStruct{
				V: "foo",
				v: []string{"bar"},
			},
			v2: NotComparableStruct{
				V: "foo",
				v: []string{"baz"},
			},
			isEqual: false,
		},
		{
			desc:    "equal floats",
			v1:      1.5,
			v2:      1.5,
			isEqual: true,
		},
		{
			desc:    "different floats",
			v1:      1.5,
			v2:      2.5,
			isEqual: false,
		},
		{
			desc:    "NaN is equal to NaN",
			v1:      math.NaN(),
			v2:      math.NaN(),
			isEqual: true,
		},
		{
			desc:    "NaN is not equal to a finite float",
			v1:      math.NaN(),
			v2:      1.0,
			isEqual: false,
		},
		{
			desc:    "float32 NaN is equal to NaN",
			v1:      float32(math.NaN()),
			v2:      float32(math.NaN()),
			isEqual: true,
		},
		{
			desc:    "named float type NaN is equal to NaN",
			v1:      NamedFloat(math.NaN()),
			v2:      NamedFloat(math.NaN()),
			isEqual: true,
		},
		{
			desc:    "+Inf is equal to +Inf",
			v1:      math.Inf(+1),
			v2:      math.Inf(+1),
			isEqual: true,
		},
		{
			desc:    "-Inf is equal to -Inf",
			v1:      math.Inf(-1),
			v2:      math.Inf(-1),
			isEqual: true,
		},
		{
			desc:    "+Inf is not equal to -Inf",
			v1:      math.Inf(+1),
			v2:      math.Inf(-1),
			isEqual: false,
		},
		{
			desc:    "NaN is not equal to +Inf",
			v1:      math.NaN(),
			v2:      math.Inf(+1),
			isEqual: false,
		},
		{
			desc:    "float slice with NaN element is equal to itself",
			v1:      []float64{1.0, 2.0, math.NaN()},
			v2:      []float64{1.0, 2.0, math.NaN()},
			isEqual: true,
		},
		{
			desc:    "float slice with NaN differs when the finite parts differ",
			v1:      []float64{1.0, 2.0, math.NaN()},
			v2:      []float64{1.0, 3.0, math.NaN()},
			isEqual: false,
		},
		{
			desc:    "float value inside a struct with a NaN field is equal",
			v1:      StructWithFloat{V: math.NaN()},
			v2:      StructWithFloat{V: math.NaN()},
			isEqual: true,
		},

		// The unexported fields of a struct passed by value can't be accessed directly,
		// yet they must be compared deeply, the same way as when the struct is passed by pointer.
		{
			desc:    "structs - unexported pointer - equal pointees behind different pointers",
			v1:      StructWithUnexportedPtr{V: "foo", p: intPtr(42)},
			v2:      StructWithUnexportedPtr{V: "foo", p: intPtr(42)},
			isEqual: true,
		},
		{
			desc:    "structs - unexported pointer - different pointees",
			v1:      StructWithUnexportedPtr{V: "foo", p: intPtr(42)},
			v2:      StructWithUnexportedPtr{V: "foo", p: intPtr(24)},
			isEqual: false,
		},
		{
			desc:    "structs - unexported pointer - nil and non-nil",
			v1:      StructWithUnexportedPtr{V: "foo"},
			v2:      StructWithUnexportedPtr{V: "foo", p: intPtr(42)},
			isEqual: false,
		},
		{
			desc:    "pointers to structs - unexported pointer - equal pointees behind different pointers",
			v1:      &StructWithUnexportedPtr{V: "foo", p: intPtr(42)},
			v2:      &StructWithUnexportedPtr{V: "foo", p: intPtr(42)},
			isEqual: true,
		},
		{
			desc:    "structs - unexported interface holding a non-comparable value - equal",
			v1:      StructWithUnexportedIface{v: []string{"foo"}},
			v2:      StructWithUnexportedIface{v: []string{"foo"}},
			isEqual: true,
		},
		{
			desc:    "structs - unexported interface holding a non-comparable value - not equal",
			v1:      StructWithUnexportedIface{v: []string{"foo"}},
			v2:      StructWithUnexportedIface{v: []string{"bar"}},
			isEqual: false,
		},
		{
			desc:    "structs - unexported non-comparable struct - equal",
			v1:      StructWithUnexportedStruct{v: NotComparableStruct{V: "foo", v: []string{"bar"}}},
			v2:      StructWithUnexportedStruct{v: NotComparableStruct{V: "foo", v: []string{"bar"}}},
			isEqual: true,
		},
		{
			desc:    "structs - unexported non-comparable struct - not equal",
			v1:      StructWithUnexportedStruct{v: NotComparableStruct{V: "foo", v: []string{"bar"}}},
			v2:      StructWithUnexportedStruct{v: NotComparableStruct{V: "foo", v: []string{"baz"}}},
			isEqual: false,
		},
		{
			desc:    "structs - unexported func - same func",
			v1:      StructWithUnexportedFunc{fn: funcA},
			v2:      StructWithUnexportedFunc{fn: funcA},
			isEqual: true,
		},
		{
			desc:    "structs - unexported func - different func",
			v1:      StructWithUnexportedFunc{fn: funcA},
			v2:      StructWithUnexportedFunc{fn: funcB},
			isEqual: false,
		},
		{
			desc:    "structs - unexported buffered channel - equal elements in different channels",
			v1:      StructWithUnexportedChan{ch: bufferedChan(1, 2)},
			v2:      StructWithUnexportedChan{ch: bufferedChan(1, 2)},
			isEqual: true,
		},
		{
			desc:    "structs - unexported buffered channel - different elements",
			v1:      StructWithUnexportedChan{ch: bufferedChan(1, 2)},
			v2:      StructWithUnexportedChan{ch: bufferedChan(1, 3)},
			isEqual: false,
		},
		{
			desc:    "structs - unexported channel - nil and non-nil are not equal",
			v1:      StructWithUnexportedChan{},
			v2:      StructWithUnexportedChan{ch: bufferedChan()},
			isEqual: false,
		},
		{
			desc:    "channels - nil and non-nil buffered are not equal",
			v1:      (chan int)(nil),
			v2:      bufferedChan(),
			isEqual: false,
		},
		{
			desc:    "channels - nil and non-nil unbuffered are not equal",
			v1:      (chan int)(nil),
			v2:      make(chan int),
			isEqual: false,
		},
		{
			desc:    "channels - closed - the buffered elements are not compared",
			v1:      closedChan(1, 2),
			v2:      closedChan(1, 3),
			isEqual: true,
		},
		{
			desc:    "channels - closed - still compared by the number of buffered elements",
			v1:      closedChan(1, 2),
			v2:      closedChan(1),
			isEqual: false,
		},

		// Structs reached by value, not through a pointer, are not addressable either.
		{
			desc:    "map values - struct with unexported pointer - equal pointees behind different pointers",
			v1:      map[string]StructWithUnexportedPtr{"k": {p: intPtr(42)}},
			v2:      map[string]StructWithUnexportedPtr{"k": {p: intPtr(42)}},
			isEqual: true,
		},
		{
			desc:    "arrays - struct with unexported pointer - equal pointees behind different pointers",
			v1:      [1]StructWithUnexportedPtr{{p: intPtr(42)}},
			v2:      [1]StructWithUnexportedPtr{{p: intPtr(42)}},
			isEqual: true,
		},
		{
			desc:    "interfaces - struct with unexported pointer - equal pointees behind different pointers",
			v1:      []any{StructWithUnexportedPtr{p: intPtr(42)}},
			v2:      []any{StructWithUnexportedPtr{p: intPtr(42)}},
			isEqual: true,
		},

		// Cycle detection must neither recurse infinitely, nor skip values that weren't compared yet.
		{
			desc:    "pointers to structs with a struct as first field - equal",
			v1:      &StructWithStructFirstField{Inner: Struct{Field1: 1}},
			v2:      &StructWithStructFirstField{Inner: Struct{Field1: 1}},
			isEqual: true,
		},
		{
			desc:    "pointers to structs with a struct as first field - not equal",
			v1:      &StructWithStructFirstField{Inner: Struct{Field1: 1}},
			v2:      &StructWithStructFirstField{Inner: Struct{Field1: 2}},
			isEqual: false,
		},
		{
			desc:    "pointers of different types to the same address - equal",
			v1:      newInteriorPointers("foo"),
			v2:      newInteriorPointers("foo"),
			isEqual: true,
		},
		{
			desc:    "pointers of different types to the same address - not equal",
			v1:      newInteriorPointers("foo"),
			v2:      newInteriorPointers("bar"),
			isEqual: false,
		},
		{
			desc:    "self-referencing structs - equal",
			v1:      newRing(1, 2, 3),
			v2:      newRing(1, 2, 3),
			isEqual: true,
		},
		{
			desc:    "self-referencing structs - not equal",
			v1:      newRing(1, 2, 3),
			v2:      newRing(1, 2, 4),
			isEqual: false,
		},
		{
			desc:    "self-referencing maps - equal",
			v1:      selfRefMap(1),
			v2:      selfRefMap(1),
			isEqual: true,
		},
		{
			desc:    "self-referencing maps - not equal",
			v1:      selfRefMap(1),
			v2:      selfRefMap(2),
			isEqual: false,
		},
		{
			desc:    "self-referencing slices - equal",
			v1:      selfRefSlice(1),
			v2:      selfRefSlice(1),
			isEqual: true,
		},
		{
			desc:    "self-referencing slices - not equal",
			v1:      selfRefSlice(1),
			v2:      selfRefSlice(2),
			isEqual: false,
		},
		{
			desc:    "the same pointer is compared against each of its counterparts",
			v1:      []*int{sharedPtr, sharedPtr},
			v2:      []*int{intPtr(42), intPtr(24)},
			isEqual: false,
		},
		{
			desc:    "slices sharing a backing array are compared separately when their lengths differ",
			v1:      StructWithSliceFields{A: backing1[:1], B: backing1[:2]},
			v2:      StructWithSliceFields{A: backing2[:1], B: backing2[:2]},
			isEqual: false,
		},
		{
			desc:    "the same buffered channel is equal to itself",
			v1:      sameChan,
			v2:      sameChan,
			isEqual: true,
		},
	}
	for _, tc := range tt {
		t.Run(tc.desc, func(t *testing.T) {
			got, err := reflects.DeepEqual(tc.v1, tc.v2)
			if !errors.Is(err, tc.hasError) {
				t.Fatalf("expected %v but got %v", tc.hasError, err)
			}
			if got != tc.isEqual {
				t.Errorf("on DeepEqual, it was expected to get %v but got %v", tc.isEqual, got)
			}
		})
	}

	t.Run("buffered channel elements are kept in their original order", func(t *testing.T) {
		// The channels are filled to capacity on purpose,
		// so putting an element back into the wrong channel would block.
		for _, elems := range [][]int{{1, 2, 3}, {1, 2, 4}} {
			ch1, ch2 := bufferedChan(1, 2, 3), bufferedChan(elems...)
			done := make(chan error, 1)
			go func() {
				_, err := reflects.DeepEqual(ch1, ch2)
				done <- err
			}()
			select {
			case err := <-done:
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
			case <-time.After(5 * time.Second):
				t.Fatalf("DeepEqual got blocked while comparing the channels")
			}
			if got := drainChan(ch1); !reflect.DeepEqual(got, []int{1, 2, 3}) {
				t.Errorf("when compared to %v, expected to keep %v, but got %v", elems, []int{1, 2, 3}, got)
			}
			if got := drainChan(ch2); !reflect.DeepEqual(got, elems) {
				t.Errorf("expected to keep %v, but got %v", elems, got)
			}
		}
	})

	t.Run("closed channels keep their buffered elements", func(t *testing.T) {
		ch1, ch2 := closedChan(1, 2, 3), bufferedChan(1, 2, 3)
		got, err := reflects.DeepEqual(ch1, ch2)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !got {
			t.Errorf("expected the channels to be equal")
		}
		if got := drainChan(ch1); !reflect.DeepEqual(got, []int{1, 2, 3}) {
			t.Errorf("expected the closed channel to keep %v, but got %v", []int{1, 2, 3}, got)
		}
		if got := drainChan(ch2); !reflect.DeepEqual(got, []int{1, 2, 3}) {
			t.Errorf("expected to keep %v, but got %v", []int{1, 2, 3}, got)
		}
	})
}

type Struct struct {
	Field1 int
	Field2 string
}

type NamedFloat float64

type StructWithFloat struct {
	V float64
}

type StructWithMethodEqual struct {
	Irrelevant int
	Relevant   int
}

func (es StructWithMethodEqual) Equal(oth StructWithMethodEqual) bool {
	return es.Relevant == oth.Relevant
}

type TestStructEquatableOnPtr struct {
	Irrelevant int
	Relevant   int
}

func (es *TestStructEquatableOnPtr) Equal(oth TestStructEquatableOnPtr) bool {
	return es.Relevant == oth.Relevant
}

type StructWithMethodEqualWithErr struct {
	Relevant int
	EqualErr error

	Irrelevant int
}

func (es StructWithMethodEqualWithErr) Equal(oth StructWithMethodEqualWithErr) (bool, error) {
	if es.EqualErr != nil {
		return false, es.EqualErr
	}
	return es.Relevant == oth.Relevant, nil
}

type StructWithMethodIsEqual struct {
	Irrelevant int
	Relevant   int
}

func (es StructWithMethodIsEqual) IsEqual(oth StructWithMethodIsEqual) bool {
	return es.Relevant == oth.Relevant
}

func cmp(a, b int) int {
	switch {
	case a < b:
		return -1
	case a == b:
		return 0
	case a > b:
		return 1
	default:
		panic("unknown Cmp case")
	}
}

type StructWithMethodCmp struct {
	Irrelevant int
	Relevant   int
}

func (es StructWithMethodCmp) Cmp(v StructWithMethodCmp) int {
	return cmp(es.Relevant, v.Relevant)
}

type StructWithCmpOnPtr struct {
	Irrelevant int
	Relevant   int
}

func (es *StructWithCmpOnPtr) Cmp(v StructWithCmpOnPtr) int {
	return cmp(es.Relevant, v.Relevant)
}

type StructWithPtrMethodCmp struct {
	Irrelevant int
	Relevant   int
}

func (es *StructWithPtrMethodCmp) Cmp(v *StructWithPtrMethodCmp) int {
	return cmp(es.Relevant, v.Relevant)
}

type ComparableStruct struct {
	V string
	v string
}

type NotComparableStruct struct {
	V string
	v []string
}

type StructWithUnexportedPtr struct {
	V string
	p *int
}

type StructWithUnexportedIface struct{ v any }

type StructWithUnexportedStruct struct{ v NotComparableStruct }

type StructWithUnexportedFunc struct{ fn func() }

func funcA() {}
func funcB() {}

type StructWithUnexportedChan struct{ ch chan int }

type StructWithStructFirstField struct {
	Inner Struct
	V     string
}

type StructWithSliceFields struct{ A, B []int }

// StructWithInteriorPointer has two pointers to the same address,
// since a struct and its first field share their address.
type StructWithInteriorPointer struct {
	Inner *Struct
	Outer *StructWithStructFirstField
}

func newInteriorPointers(v string) StructWithInteriorPointer {
	outer := &StructWithStructFirstField{Inner: Struct{Field1: 1}, V: v}
	return StructWithInteriorPointer{Inner: &outer.Inner, Outer: outer}
}

type RingNode struct {
	V    int
	next *RingNode
}

// newRing returns the first node of a ring, where the last node points back to the first.
func newRing(values ...int) RingNode {
	nodes := make([]*RingNode, len(values))
	for i, v := range values {
		nodes[i] = &RingNode{V: v}
	}
	for i := range nodes {
		nodes[i].next = nodes[(i+1)%len(nodes)]
	}
	return *nodes[0]
}

func selfRefMap(v int) map[string]any {
	m := map[string]any{"v": v}
	m["self"] = m
	return m
}

func selfRefSlice(v int) []any {
	s := []any{v, nil}
	s[1] = s
	return s
}

func intPtr(n int) *int { return &n }

const bufferedChanCap = 3

func bufferedChan(vs ...int) chan int {
	ch := make(chan int, bufferedChanCap)
	for _, v := range vs {
		ch <- v
	}
	return ch
}

func closedChan(vs ...int) chan int {
	ch := bufferedChan(vs...)
	close(ch)
	return ch
}

func drainChan(ch chan int) []int {
	var vs []int
	for n := len(ch); len(vs) < n; {
		vs = append(vs, <-ch)
	}
	return vs
}

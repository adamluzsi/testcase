package reflects_test

import (
	"errors"
	"reflect"
	"testing"

	"go.llib.dev/testcase/internal/reflects"
	"go.llib.dev/testcase/random"
)

func TestDeepEqual(t *testing.T) {
	rnd := random.New(random.CryptoSeed{})
	expErr := rnd.Error()
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
}

type Struct struct {
	Field1 int
	Field2 string
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

func TestDeepEqual_reflectType(t *testing.T) {
	var (
		v1 = reflect.TypeOf((*string)(nil)).Elem()
		v2 = reflect.TypeOf((*int)(nil)).Elem()
	)
	isEqual, err := reflects.DeepEqual(v1, v2)
	if err != nil {
		t.Fatal(err.Error())
	}
	if isEqual {
		t.Fatalf("unexpected equality between two reflect.Type value")
	}
}

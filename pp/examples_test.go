package pp_test

import (
	"github.com/adamluzsi/testcase/pp"
)

type ExampleStruct struct {
	A string
	B int
}

func ExampleFormat() {
	_ = pp.Format(ExampleStruct{
		A: "The Answer",
		B: 42,
	})
}

func ExampleDiff() {
	_ = pp.Diff(ExampleStruct{
		A: "The Answer",
		B: 42,
	}, ExampleStruct{
		A: "The Question",
		B: 42,
	})
}

func ExampleDiffString() {
	_ = pp.Diff("aaa\nbbb\nccc\n", "aaa\nccc\n")
}

package pp_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"go.llib.dev/testcase/pp"
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
	pp.DiffFormat(ExampleStruct{
		A: "The Answer",
		B: 42,
	}, ExampleStruct{
		A: "The Question",
		B: 42,
	})
}

func ExampleDiffFormat() {
	fmt.Println(pp.DiffFormat(ExampleStruct{
		A: "The Answer",
		B: 42,
	}, ExampleStruct{
		A: "The Question",
		B: 42,
	}))
}

func ExampleDiffString() {
	_ = pp.DiffFormat("aaa\nbbb\nccc\n", "aaa\nccc\n")
}

func ExamplePP_unexportedFields() {
	var buf bytes.Buffer
	bs, _ := json.Marshal(ExampleStruct{
		A: "The Answer",
		B: 42,
	})
	buf.Write(bs)

	pp.PP(buf)
}

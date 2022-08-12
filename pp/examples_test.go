package pp_test

import (
	"bytes"
	"encoding/json"
	"github.com/adamluzsi/testcase/pp"
	"testing"
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

func ExamplePP_unexportedFields() {
	var buf bytes.Buffer
	bs, _ := json.Marshal(ExampleStruct{
		A: "The Answer",
		B: 42,
	})
	buf.Write(bs)

	pp.PP(buf)
}

func TestXXX(t *testing.T) {
	println(pp.Diff(ExampleStruct{
		A: "The Answer",
		B: 42,
	}, ExampleStruct{
		A: "The Question",
		B: 42,
	}))
}

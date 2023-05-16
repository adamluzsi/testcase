package pp

import (
	"bufio"
	"bytes"
	"strings"
	"testing"
)

const DiffOutput = `
pp.X{       pp.X{
  A: 1,  |    A: 2,
  B: 2,       B: 2,
}           }
`

func TestDiff_smoke(t *testing.T) {
	ogw := defaultWriter
	defer func() { defaultWriter = ogw }()
	buf := &bytes.Buffer{}
	defaultWriter = buf

	type X struct{ A, B int }
	v1 := X{A: 1, B: 2}
	v2 := X{A: 2, B: 2}
	Diff(v1, v2)

	exp := strings.TrimSpace(DiffOutput)
	got := strings.TrimSpace(buf.String())
	mustEqual(t, exp, got)
}

func TestDiffFormat_smoke(t *testing.T) {
	type X struct{ A, B int }
	v1 := X{A: 1, B: 2}
	v2 := X{A: 2, B: 2}
	tr := strings.TrimSpace
	mustEqual(t, tr(DiffOutput), tr(DiffFormat(v1, v2)))
}

const DiffStringA = `
aaa
bbb
ccc
ddd
eee
fff
ggg
`

const DiffStringB = `
aaa
bbbdiff
ccc
eee
123
fff
`

const DiffStringOut = `
aaa     aaa
bbb  |  bbbdiff
ccc     ccc
ddd  <  
eee     eee
     >  123
fff     fff
ggg  <
`

func TestPrettyPrinter_DiffString_smoke(t *testing.T) {
	t.Run("E2E", func(t *testing.T) {
		tr := strings.TrimSpace
		got := DiffString(tr(DiffStringA), tr(DiffStringB))
		t.Logf("\n%s", got)
		exp := tr(DiffStringOut)
		act := tr(got)
		t.Logf("\n\nexpected:\n%s\n\nactual:\n%s", exp, act)
		mustEqual(t, exp, act)
	})
	tr := func(str string) string {
		var strs []string
		s := bufio.NewScanner(strings.NewReader(str))
		s.Split(bufio.ScanLines)
		for s.Scan() {
			strs = append(strs, strings.TrimSpace(s.Text()))
		}
		return strings.Join(strs, "\n")
	}
	type TestCase struct {
		Desc string
		A    string
		B    string
		Diff string
	}
	for _, tc := range []TestCase{
		{
			Desc: "when only A has value",
			A:    "aaa",
			B:    "",
			Diff: "aaa  <",
		},
		{
			Desc: "when only B has value",
			A:    "",
			B:    "bbb",
			Diff: ">  bbb",
		},
		{
			Desc: "when A and B not equals",
			A:    "aaa",
			B:    "bbb",
			Diff: "aaa  |  bbb",
		},
		{
			Desc: "when A has values as B plus more in the middle",
			A:    "aaa\n123\nbbb",
			B:    "aaa\nbbb\n",
			Diff: "aaa     aaa\n123  <\nbbb     bbb",
		},
		{
			Desc: "when B has values as A plus more in the middle",
			A:    "aaa\nbbb",
			B:    "aaa\n123\nbbb\n",
			Diff: "aaa     aaa\n>  123\nbbb     bbb",
		},
		{
			Desc: "when A has values as B plus more afterwards",
			A:    "aaa\nbbb\n123",
			B:    "aaa\nbbb\n",
			Diff: "aaa     aaa\nbbb     bbb\n123  <",
		},
		{
			Desc: "when B has values as A plus more afterwards",
			A:    "aaa\nbbb\n",
			B:    "aaa\nbbb\n123",
			Diff: "aaa     aaa\nbbb     bbb\n>  123",
		},
	} {
		tc := tc
		t.Run(tc.Desc, func(t *testing.T) {
			diff := DiffString(tr(tc.A), tr(tc.B))
			mustEqual(t, tr(tc.Diff), tr(diff))
		})
	}
}

func mustEqual(tb testing.TB, exp string, act string) {
	tb.Helper()
	if act != exp {
		tb.Fatalf("exp and got not equal: \n\nexpected:\n%s\n\nactual:\n%s\n", exp, act)
	}
}

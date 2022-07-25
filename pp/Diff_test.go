package pp_test

import (
	"bufio"
	"strings"
	"testing"

	"github.com/adamluzsi/testcase/assert"
	"github.com/adamluzsi/testcase/pp"
)

const DiffOutput = `
pp_test.X{     pp_test.X{
  A: 1,     |    A: 2,
  B: 2,          B: 2,
}              }
`

func TestDiff_smoke(t *testing.T) {
	type X struct{ A, B int }
	v1 := X{A: 1, B: 2}
	v2 := X{A: 2, B: 2}
	tr := strings.TrimSpace
	assert.Equal(t, tr(DiffOutput), tr(pp.Diff(v1, v2)))
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
		got := pp.DiffString(tr(DiffStringA), tr(DiffStringB))
		t.Logf("\n%s", got)
		exp := tr(DiffStringOut)
		act := tr(got)
		t.Logf("\n\nexpected:\n%s\n\nactual:\n%s", exp, act)
		assert.Equal(t, exp, act)
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
			diff := pp.DiffString(tr(tc.A), tr(tc.B))
			assert.Equal(t, tr(tc.Diff), tr(diff))
		})
	}
}

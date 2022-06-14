package source_test

import (
	"bytes"
	"go/ast"
	"testing"

	"github.com/adamluzsi/testcase/assert"
	"github.com/adamluzsi/testcase/internal/source"
)

const ExampleTestFilePath = "./fixtures/example_test.go"

func Must1[T any](v T, err error) T {
	if err != nil {
		panic(err.Error())
	}
	return v
}

func TestParser_smoke(t *testing.T) {
	it := assert.MakeIt(t)
	sp := source.Parser{}
	sp.Init()

	it.Must.NoError(sp.ParseFile(ExampleTestFilePath))

	list, ok := sp.FindTCRBlockByPosition(ExampleTestFilePath, 38)
	assert.True(t, ok)
	assert.Equal(t, 1, len(list))

	anyOf := assert.AnyOf{TB: t, Fn: t.Fatal}
	for _, n := range list {
		anyOf.Test(func(it assert.It) { it.Must.Contain(Must1(sp.SourceString(n)), "bar()") })
	}
	anyOf.Finish()

	buf := &bytes.Buffer{}
	sp.FprintSource(buf, &ast.BlockStmt{
		Lbrace: 0,
		List:   list,
		Rbrace: 0,
	})
	assert.Contain(t, buf.String(), "bar()")

	//list, ok = sp.FindTCRBlockByPosition(ExampleTestFilePath, 39)
	//assert.True(t, ok)
	//
	//fmt.Println(len(list))
	//buf = &bytes.Buffer{}
	//assert.Contain(t, buf.String(), "bar()")
}

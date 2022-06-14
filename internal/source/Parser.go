package source

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/format"
	"go/parser"
	"go/token"
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/adamluzsi/testcase/internal"
)

type Parser struct {
	FilePath string

	init sync.Once

	FileSet          *token.FileSet
	PkgImportMapping map[ /*ident*/ string] /*path*/ string
	AstFile          *ast.File
	Packages         map[string]*ast.Package
}

func (sp *Parser) Init() {
	sp.init.Do(func() {
		sp.FileSet = token.NewFileSet()
		sp.PkgImportMapping = make(map[string]string)
		sp.Packages = make(map[string]*ast.Package)
	})
}

func (sp *Parser) ParseFile(fpath string) error {
	sp.Init()

	for _, pkg := range sp.Packages {
		for _, f := range pkg.Files {
			if f.Name.Name == fpath {
				fmt.Println(f.Name.Name, fpath)
				return nil
			}
		}
	}

	pkgs, err := parser.ParseDir(sp.FileSet, filepath.Dir(fpath), nil, parser.ParseComments)
	if err != nil {
		return err
	}
	for name, pkg := range pkgs { // merge
		sp.Packages[name] = pkg
	}

	// refresh importPath mapping
	sp.Inspect(func(node ast.Node) bool {
		switch node := node.(type) {
		case *ast.ImportSpec:
			importPath := strings.Trim(node.Path.Value, `"`+"`")
			var name = filepath.Base(importPath)
			if node.Name != nil {
				name = node.Name.Name
			}
			sp.PkgImportMapping[name] = importPath
		}
		return true
	})
	return nil
}

type VCTRL struct {
	Current ast.Node
	File    *ast.File
}

var (
	stopVisit     = fmt.Errorf("VCTRL:Stop")
	stepOverVisit = fmt.Errorf("VCTRL:StepOver")
)

func (VCTRL) Stop() {
	panic(stopVisit)
}

func (VCTRL) StepOver() {
	panic(stepOverVisit)
}

func (sp *Parser) Visit(fn func(VCTRL)) {
	for _, pkg := range sp.Packages {
		var stop bool
		var lastFile *ast.File
		ast.Inspect(pkg, func(node ast.Node) bool {
			if node == nil || stop {
				lastFile = nil
				return false
			}
			if file, ok := node.(*ast.File); ok {
				lastFile = file
				return true
			}

			var recovered any
			func() {
				defer func() { recovered = recover() }()
				fn(VCTRL{
					Current: node,
					File:    lastFile,
				})
			}()

			switch recovered {
			case stopVisit:
				stop = true
				return false
			case stepOverVisit:
				return false
			}

			return true
		})
		if stop {
			break
		}
	}
}

func (sp *Parser) Inspect(fn func(ast.Node) bool) {
	for _, pkg := range sp.Packages {
		ast.Inspect(pkg, fn)
	}
}

func (sp *Parser) FprintSource(w io.Writer, node ast.Node) error {
	return format.Node(w, sp.FileSet, node)
}

func (sp *Parser) PrintSource(node ast.Node) error {
	return sp.FprintSource(os.Stdout, node)
}

func (sp *Parser) SourceString(node ast.Node) (string, error) {
	out := &bytes.Buffer{}
	if err := sp.FprintSource(out, node); err != nil {
		return "", err
	}
	return out.String(), nil
}

func (sp *Parser) PrintAst(node ast.Node) {
	ast.Print(sp.FileSet, node)
}

// AssertFuncLitRuntimeBlock
//
// 0: *ast.FuncLit {
// .  Type: *ast.FuncType {
// .  .  Func: ./fixtures/example_test.go:38:12
// .  .  Params: *ast.FieldList {
// .  .  .  Opening: ./fixtures/example_test.go:38:16
// .  .  .  List: []*ast.Field (len = 1) {
// .  .  .  .  0: *ast.Field {
// .  .  .  .  .  Names: []*ast.Ident (len = 1) {
// .  .  .  .  .  .  0: *ast.Ident {
// .  .  .  .  .  .  .  NamePos: ./fixtures/example_test.go:38:17
// .  .  .  .  .  .  .  Name: "t"
// .  .  .  .  .  .  .  Obj: *ast.Object {
// .  .  .  .  .  .  .  .  Kind: var
// .  .  .  .  .  .  .  .  Name: "t"
// .  .  .  .  .  .  .  .  Decl: *(obj @ 908)
// .  .  .  .  .  .  .  }
// .  .  .  .  .  .  }
// .  .  .  .  .  }
// .  .  .  .  .  Type: *ast.StarExpr {
// .  .  .  .  .  .  Star: ./fixtures/example_test.go:38:19
// .  .  .  .  .  .  X: *ast.SelectorExpr {
// .  .  .  .  .  .  .  X: *ast.Ident {
// .  .  .  .  .  .  .  .  NamePos: ./fixtures/example_test.go:38:20
// .  .  .  .  .  .  .  .  Name: "tc"
// .  .  .  .  .  .  .  }
// .  .  .  .  .  .  .  Sel: *ast.Ident {
// .  .  .  .  .  .  .  .  NamePos: ./fixtures/example_test.go:38:23
// .  .  .  .  .  .  .  .  Name: "T"
// .  .  .  .  .  .  .  }
// .  .  .  .  .  .  }
// .  .  .  .  .  }
// .  .  .  .  }
// .  .  .  }
// .  .  .  Closing: ./fixtures/example_test.go:38:24
// .  .  }
// .  }
// .  Body: *ast.BlockStmt {
// .  .  Lbrace: ./fixtures/example_test.go:38:26
// .  .  List: []ast.Stmt (len = 1) {
// .  .  .  0: *ast.ExprStmt {
// .  .  .  .  X: *ast.CallExpr {
// .  .  .  .  .  Fun: *ast.Ident {
// .  .  .  .  .  .  NamePos: ./fixtures/example_test.go:39:4
// .  .  .  .  .  .  Name: "world"
// .  .  .  .  .  .  Obj: *(obj @ 479)
// .  .  .  .  .  }
// .  .  .  .  .  Lparen: ./fixtures/example_test.go:39:9
// .  .  .  .  .  Ellipsis: -
// .  .  .  .  .  Rparen: ./fixtures/example_test.go:39:10
// .  .  .  .  }
// .  .  .  }
// .  .  }
// .  .  Rbrace: ./fixtures/example_test.go:40:3
// .  }
func (sp Parser) AssertFuncLitRuntimeBlock(node ast.Node) (_blk []ast.Stmt, _ok bool) {
	fl, ok := node.(*ast.FuncLit)
	if !ok ||
		fl == nil ||
		fl.Type == nil ||
		fl.Type.Params == nil {
		return
	}
	if len(fl.Type.Params.List) < 1 || fl.Type.Params.List[0] == nil {
		// not enough parameter for a testcase runtime function body
		return
	}
	fnBlocktestcaseTArg := fl.Type.Params.List[0]

	isTestcaseT := func(node *ast.Field) bool {
		se, ok := node.Type.(*ast.StarExpr)
		if !ok || se == nil {
			return false
		}
		sel, ok := se.X.(*ast.SelectorExpr)
		if !ok || sel == nil {
			return false
		}

		// check testcase.T
		if sel.Sel.Name != "T" {
			return true
		}
		ident, ok := sel.X.(*ast.Ident)
		if !ok || ident == nil {
			return false
		}
		if pkg, ok := sp.PkgImportMapping[ident.Name]; !ok || !strings.HasSuffix(pkg, "testcase") {
			return false
		}
		return true
	}

	if !isTestcaseT(fnBlocktestcaseTArg) {
		return
	}

	return fl.Body.List, true
}

func (sp *Parser) FindTCRBlockByPosition(file string, line int) (_rblk []ast.Stmt, _ok bool) {
	var stack internal.Stack[ast.Node]
	sp.Inspect(func(node ast.Node) (cont bool) {
		defer func() {
			if !cont {
				stack.Pop()
			}
		}()
		if node == nil || _ok {
			return false
		}
		stack.Push(node)

		pos := sp.FileSet.Position(node.Pos())

		if !strings.Contains(file, pos.Filename) {
			return false
		}
		if pos.Line != line {
			return true
		}

		for i := len(stack) - 1; i >= 0; i-- {
			if block, ok := sp.AssertFuncLitRuntimeBlock(stack[i]); ok {
				_ok = true
				_rblk = block
				return false
			}
		}

		return true
	})
	if len(stack) != 0 {
		panic(fmt.Sprintf("stack is in a dirty state! len:%d", len(stack)))
	}
	return
}

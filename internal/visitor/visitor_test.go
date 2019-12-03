package visitor_test

import (
	"fmt"
	"go/ast"
	"go/token"
	"io/ioutil"
	"os"
	"path"
	"testing"

	"github.com/romanyx/erris/internal/visitor"
	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/packages"
	"golang.org/x/tools/imports"
)

type position struct{ Line int }

func (p position) Pos() token.Pos { return token.Pos(p.Line) }
func (p position) End() token.Pos { return token.NoPos }

func TestVisitorVisit(t *testing.T) {
	tt := []struct {
		name    string
		content string
		expect  visitor.Issues
	}{
		{
			name: "without issues",
			content: `
			x := 1
			if x == 2 {}
			err := errors.New("mock error")
			if errors.Is(err, io.EOF) {}
			var r io.Reader
			if errors.As(err, &r) {}
			`,
			expect: make(visitor.Issues, 0),
		},
		{
			name: "equasion",
			content: `
			err := errors.New("mock error")
			if err == io.EOF {}
			if err != io.EOF {}
			`,
			expect: visitor.Issues{
				{
					Text: "use errors.Is to compare an error",
					Node: position{
						Line: 11,
					},
				},
				{
					Text: "use errors.Is to compare an error",
					Node: position{
						Line: 13,
					},
				},
			},
		},
		{
			name: "assertion",
			content: `
			err := errors.New("mock error")
			if _, ok := err.(io.Reader); ok {}
			switch err.(type) {
			case io.Reader:
			}
			`,
			expect: visitor.Issues{
				{
					Text: "use errors.As to assert an error",
					Node: position{
						Line: 11,
					},
				},
				{
					Text: "use errors.As to assert an error",
					Node: position{
						Line: 13,
					},
				},
			},
		},
	}

	for _, tc := range tt {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			pass, err := visitorSuite(t, fmt.Sprintf(baseContent, tc.content))
			if err != nil {
				t.Error(err)
				return
			}

			v := visitor.New(pass)

			for _, astFile := range pass.Files {
				ast.Walk(v, astFile)
			}

			assertIssues(t, pass.Fset, v.Issues, tc.expect)
		})
	}
}

func assertIssues(t *testing.T, fset *token.FileSet, got, expect visitor.Issues) {
	t.Helper()

	if len(got) != len(expect) {
		t.Errorf("got %d and expect %d has different lengths", len(got), len(expect))
		return
	}

	for i, issue := range got {
		if issue.Text != expect[i].Text {
			t.Errorf("got text: '%s' expected: '%s'", issue.Text, expect[i].Text)
		}

		if fset.Position(issue.Node.Pos()).Line != expect[i].Node.(position).Line {
			t.Errorf("got pos: %d expected: %d", issue.Node.Pos(), expect[i].Node.Pos())
		}
	}
}

func visitorSuite(t *testing.T, content string) (*analysis.Pass, error) {
	t.Helper()

	// copy testvendor directory into directory for test
	tmpGopath, err := ioutil.TempDir("", "testvendor")
	if err != nil {
		return nil, fmt.Errorf("unable to create testvendor directory: %v", err)
	}
	testVendorDir := path.Join(tmpGopath, "src", "github.com/testvendor")
	if err := os.MkdirAll(testVendorDir, 0755); err != nil {
		return nil, fmt.Errorf("mkdir all failed: %v", err)
	}
	defer func() {
		os.RemoveAll(tmpGopath)
	}()

	// Format code using goimport standard
	path := path.Join(testVendorDir, "main.go")
	bs, err := imports.Process(path, []byte(content), nil)
	if err != nil {
		return nil, fmt.Errorf("process imports: %v", err)
	}
	if err := ioutil.WriteFile(path, bs, 0755); err != nil {
		return nil, fmt.Errorf("failed to write testvendor main: %v", err)
	}

	cfg := &packages.Config{
		Mode: packages.LoadAllSyntax,
		Env:  append(os.Environ(), "GOPATH="+tmpGopath),
		Dir:  testVendorDir,
	}

	pkgs, err := packages.Load(cfg, "github.com/testvendor")
	if err != nil {
		return nil, fmt.Errorf("load packages: %v", err)
	}
	if len(pkgs) != 1 {
		return nil, fmt.Errorf("got more than one package")
	}
	pkg := pkgs[0]
	if len(pkg.Errors) > 0 {
		return nil, fmt.Errorf(
			"errors while loading package %s: %v",
			pkg.ID,
			pkg.Errors,
		)
	}
	return &analysis.Pass{
		Fset:      pkg.Fset,
		Files:     pkg.Syntax,
		TypesInfo: pkg.TypesInfo,
	}, nil
}

var baseContent = `package main

func main() {
	%s
}
`

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
	"golang.org/x/tools/go/packages"
	"golang.org/x/tools/imports"
)

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
					Pos: token.Position{
						Line: 11,
					},
				},
				{
					Text: "use errors.Is to compare an error",
					Pos: token.Position{
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
					Pos: token.Position{
						Line: 11,
					},
				},
				{
					Text: "use errors.As to assert an error",
					Pos: token.Position{
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
			pkgs, err := visitorSuite(t, fmt.Sprintf(baseContent, tc.content))
			if err != nil {
				t.Error(err)
				return
			}

			issues := make(visitor.Issues, 0)
			for _, pkg := range pkgs {
				v := visitor.New(pkg)

				for _, astFile := range pkg.Syntax {
					ast.Walk(v, astFile)
				}
				issues = append(issues, v.Issues...)
			}

			assertIssues(t, issues, tc.expect)
		})
	}
}

func assertIssues(t *testing.T, got, expect visitor.Issues) {
	t.Helper()

	if len(got) != len(expect) {
		t.Errorf("got %d and expect %d has different lengths", len(got), len(expect))
		return
	}

	for i, issue := range got {
		if issue.Text != expect[i].Text {
			t.Errorf("got text: '%s' expected: '%s'", issue.Text, expect[i].Text)
		}

		if issue.Pos.Line != expect[i].Pos.Line {
			t.Errorf("got line: %d expected: %d", issue.Pos.Line, expect[i].Pos.Line)
		}
	}
}

func visitorSuite(t *testing.T, content string) ([]*packages.Package, error) {
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

	for _, pkg := range pkgs {
		if len(pkg.Errors) > 0 {
			return nil, fmt.Errorf(
				"errors while loading package %s: %v",
				pkg.ID,
				pkg.Errors,
			)
		}
	}

	return pkgs, nil
}

var baseContent = `package main

func main() {
	%s
}
`

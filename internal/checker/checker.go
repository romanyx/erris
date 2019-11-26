package checker

import (
	"fmt"
	"go/ast"

	"golang.org/x/tools/go/packages"

	"github.com/romanyx/erris/internal/visitor"
)

// New prepares checker.
func New(withoutTests bool) *Checker {
	c := Checker{
		withoutTests: withoutTests,
	}

	return &c
}

// Checker allows to check paths.
type Checker struct {
	withoutTests bool
}

// CheckPackages checks packages for issues.
func (c *Checker) CheckPackages(paths ...string) error {
	pkgs, err := c.load(paths...)
	if err != nil {
		return err
	}

	// Check for errors in the initial packages.
	for _, pkg := range pkgs {
		if len(pkg.Errors) > 0 {
			return fmt.Errorf(
				"errors while loading package %s: %v",
				pkg.ID,
				pkg.Errors,
			)
		}
	}

	issues := make(visitor.Issues, 0, 20)
	for _, pkg := range pkgs {
		v := visitor.New(pkg)

		for _, astFile := range pkg.Syntax {
			ast.Walk(v, astFile)
		}

		issues = append(issues, v.Issues...)
	}

	if len(issues) > 0 {
		uniq := issues[:0] // compact in-place
		for i, err := range issues {
			if i == 0 || err != issues[i-1] {
				uniq = append(uniq, err)
			}
		}

		return uniq
	}

	return nil
}

func (c *Checker) load(paths ...string) ([]*packages.Package, error) {
	cfg := &packages.Config{
		Mode:  packages.LoadAllSyntax,
		Tests: !c.withoutTests,
	}

	return packages.Load(cfg, paths...)

}

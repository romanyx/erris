package visitor

import (
	"go/ast"
	"go/token"
	"go/types"

	"golang.org/x/tools/go/packages"
)

// New returns prepared visitor.
func New(pkg *packages.Package) *Visitor {
	v := Visitor{
		errorType: types.Universe.
			Lookup("error").
			Type().
			Underlying().(*types.Interface),
		pkg:    pkg,
		Issues: make(Issues, 0, 10),
	}

	return &v
}

// Visitor holds package and issues.
type Visitor struct {
	errorType *types.Interface
	pkg       *packages.Package
	Issues    []Issue
}

// Issue representation.
type Issue struct {
	Text string
	Pos  token.Position
}

// Issues holder for issue.
type Issues []Issue

// Error implements error.
func (i Issues) Error() string {
	return "erris issues found"
}

// Visit implements ast.Visitor.
func (v *Visitor) Visit(node ast.Node) ast.Visitor {
	if node == nil {
		return v
	}

	switch t := node.(type) {
	// if err == sql.ErrNoRows
	// return err == sql.ErrNoRows
	case *ast.BinaryExpr:
		if t.Op != token.EQL && t.Op != token.NEQ {
			break
		}

		x, y := v.typeOf(t.X), v.typeOf(t.Y)
		if v.isErrorTypes(x, y) {
			v.Issues = append(v.Issues, Issue{
				Text: "use errors.Is to compare an error",
				Pos:  v.position(t),
			})
		}
	// switch err.(type)
	// _, ok := err.(T)
	case *ast.TypeAssertExpr:
		x := v.typeOf(t.X)
		if v.isErrorTypes(x) {
			v.Issues = append(v.Issues, Issue{
				Text: "use errors.As to assert an error",
				Pos:  v.position(t),
			})
		}
	}

	return v
}

func (v *Visitor) isErrorTypes(ts ...types.Type) bool {
	for _, t := range ts {
		if !types.Implements(t, v.errorType) {
			return false
		}
	}

	return true
}

func (v *Visitor) typeOf(e ast.Expr) types.Type {
	return v.pkg.TypesInfo.TypeOf(e)
}

func (v *Visitor) position(n ast.Node) (pos token.Position) {
	return v.pkg.Fset.Position(n.Pos())
}

package erris

import (
	"go/ast"

	"golang.org/x/tools/go/analysis"

	"github.com/romanyx/erris/internal/visitor"
)

var _ = Analyzer.Flags.Bool("ignoretests", false, "this flag is deprecated and has no effect")

// Analyzer is an golang.org/x/tools/go/analysis compatible interface.
var Analyzer = &analysis.Analyzer{
	Name:             "erris",
	Doc:              `checks that errors are compared or type asserted using errors.Is and errors.As`,
	Run:              run,
	RunDespiteErrors: true,
}

func run(p *analysis.Pass) (interface{}, error) {
	v := visitor.New(p)
	for _, f := range p.Files {
		ast.Walk(v, f)
	}
	for _, i := range v.Issues {
		p.ReportRangef(i.Node, i.Text)
	}
	return nil, nil
}

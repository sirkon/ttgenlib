package predictor

import (
	"go/ast"

	"github.com/posener/complete"
)

// FunctionName completes function name of the given package.
type FunctionName struct{}

// Predict to satisfy complete.Predictor.
func (f FunctionName) Predict(args complete.Args) []string {
	dir := getPackageContent(args.All)
	if dir == nil {
		return nil
	}

	var funcs []string
	for _, pkg := range dir {
		for _, file := range pkg.Files {
			ast.Inspect(file, func(node ast.Node) bool {
				v, ok := node.(*ast.FuncDecl)
				if !ok {
					return false
				}

				if v.Recv != nil && len(v.Recv.List) > 0 {
					return false
				}

				funcs = append(funcs, v.Name.Name)
				return false
			})
		}
	}

	return funcs
}

var _ complete.Predictor = FunctionName{}

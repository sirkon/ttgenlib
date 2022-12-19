package predictor

import (
	"go/ast"

	"github.com/posener/complete"
)

// TypeMethodName complete method names for a type.
type TypeMethodName struct{}

// Predict to satisfy complete.Predictor.
func (TypeMethodName) Predict(args complete.Args) []string {
	dirs := getPackageContent(args.All)
	if dirs == nil {
		return nil
	}

	if len(args.Completed) == 0 {
		return nil
	}

	tn := args.Completed[len(args.Completed)-1]
	stn := "*" + tn
	var methodNames []string
	for _, dir := range dirs {
		for _, file := range dir.Files {
			ast.Inspect(file, func(node ast.Node) bool {
				v, ok := node.(*ast.FuncDecl)
				if !ok {
					return false
				}

				if v.Recv == nil || len(v.Recv.List) == 0 {
					return false
				}

				recv := v.Recv.List[0]
				switch typeName(recv.Type) {
				case tn, stn:
					methodNames = append(methodNames, v.Name.Name)
				}

				return false
			})
		}
	}

	return methodNames
}

func typeName(te ast.Expr) string {
	switch v := te.(type) {
	case *ast.StarExpr:
		return "*" + typeName(v.X)
	case *ast.Ident:
		return v.Name
	default:
		return ""
	}
}

var _ complete.Predictor = TypeMethodName{}

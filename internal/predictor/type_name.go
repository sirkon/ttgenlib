package predictor

import (
	"go/ast"

	"github.com/posener/complete"
)

// TypeName complete types names.
type TypeName struct{}

// Predict to satisfy complete.Predictor.
func (TypeName) Predict(args complete.Args) []string {
	dirs := getPackageContent(args.All)
	if dirs == nil {
		return nil
	}

	var typeNames []string
	for _, dir := range dirs {
		for _, file := range dir.Files {
			ast.Inspect(file, func(node ast.Node) bool {
				v, ok := node.(*ast.TypeSpec)
				if !ok {
					return false
				}

				_, ok = v.Type.(*ast.StructType)
				if !ok {
					return false
				}

				typeNames = append(typeNames, v.Name.Name)
				return false
			})
		}
	}

	return typeNames
}

var _ complete.Predictor = TypeName{}

package predictor

import (
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"strings"

	"github.com/sirkon/jsonexec"
)

func getPackagePath(args []string) string {
	for i := 0; i < len(args); i++ {
		if args[i] == "-p" || args[i] == "--pkg-path" {
			if i < len(args)-1 {
				return args[i+1]
			}
		}
	}

	return ""
}

func fileFilter(info os.FileInfo) bool {
	return strings.HasSuffix(info.Name(), ".go") && !strings.HasSuffix(info.Name(), "_test.go")
}

func getPackageContent(args []string) map[string]*ast.Package {
	pkgpath := getPackagePath(args)
	if pkgpath == "" {
		return nil
	}

	var dst struct {
		Dir string
	}
	if err := jsonexec.Run(&dst, "go", "list", "--json", pkgpath); err != nil {
		return nil
	}

	dir, err := parser.ParseDir(token.NewFileSet(), dst.Dir, fileFilter, parsingMode)
	if err != nil {
		return nil
	}

	return dir
}

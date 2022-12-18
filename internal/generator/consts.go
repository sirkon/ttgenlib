package generator

import "golang.org/x/tools/go/packages"

var (
	contextNoMock = doNotMock{
		path: "context",
		name: "Context",
	}
)

const (
	gomockPath       = "github.com/golang/mock/gomock"
	gomockController = "Controller"
	deepequalPath    = "github.com/sirkon/deepequal"

	PackageLoadMode = packages.NeedImports | packages.NeedTypes | packages.NeedName | packages.NeedDeps |
		packages.NeedSyntax | packages.NeedFiles | packages.NeedModule | packages.NeedSyntax
)

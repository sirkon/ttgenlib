package generator

import "github.com/sirkon/gogh"

type (
	goPackage  = gogh.Package[*gogh.Imports]
	goRenderer = gogh.GoRenderer[*gogh.Imports]
)

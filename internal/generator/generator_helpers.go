package generator

import (
	"go/token"
	"go/types"
	"path"

	"github.com/sirkon/message"
)

func (g *Generator) digObjectFile(f types.Object) string {
	_, file := path.Split(g.fset.Position(f.Pos()).Filename)
	return file
}

func (g *Generator) shouldNotBeMocked(vn *types.Named) bool {
	if vn.Obj().Pkg() == nil {
		return true
	}

	for _, c := range g.nomock {
		if vn.Obj().Name() == c.name && (vn.Obj().Pkg() != nil && vn.Obj().Pkg().Path() == c.path) {
			return true
		}
	}

	return false
}

func (g *Generator) infoParamNotInterfaceOmit(pos token.Pos, name string) {
	message.Debugf("%s type of parameter %s is not an interface, omitting", g.fset.Position(pos), name)
}

func (g *Generator) infoFieldNotInterfaceOmit(pos token.Pos, name string) {
	message.Debugf("%s type of field %s is not an interface, omitting", g.fset.Position(pos), name)
}

func underlyingTypeIs[T types.Type](v *types.Named) bool {
	_, ok := v.Underlying().(T)
	return ok
}

func isContext(v types.Type) bool {
	n, ok := v.(*types.Named)
	if !ok {
		return false
	}

	return n.Obj().Pkg().Path() == "context" && n.Obj().Name() == "Context"
}

func hasContextArg(s *types.Signature) bool {
	for i := 0; i < s.Params().Len(); i++ {
		if isContext(s.Params().At(i).Type()) {
			return true
		}
	}

	return false
}

func isError(v types.Type) bool {
	vn, ok := v.(*types.Named)
	if !ok {
		return false
	}

	if vn.Obj().Name() != "error" {
		return false
	}

	iface, ok := vn.Underlying().(*types.Interface)
	if !ok {
		return false
	}

	if iface.NumMethods() != 1 {
		return false
	}

	m := iface.Method(0)
	if m.Name() != "Error" {
		return false
	}

	s := m.Type().(*types.Signature)
	if s.Params().Len() != 0 {
		return false
	}

	if s.Results().Len() != 1 {
		return false
	}

	if s.Results().At(0).Type().String() != "string" {
		return false
	}

	return true
}

func isErrored(s *types.Signature) bool {
	if s.Results().Len() == 0 {
		return false
	}

	return isError(s.Results().At(s.Results().Len() - 1).Type())
}

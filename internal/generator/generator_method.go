package generator

import (
	"go/types"
	"strings"

	"github.com/sirkon/errors"
)

// GenerateForMethod generates table test template and helpers for a method of a type.
func GenerateForMethod(pkg, typ, method string, mockLookup MockLookup, opts ...Option) error {
	g, err := newGenerator(pkg, mockLookup, opts...)
	if err != nil {
		return errors.Wrap(err, "init generator")
	}

	t := g.pkg.Types.Scope().Lookup(typ)
	if t == nil {
		return errors.Newf("type %s not found", typ)
	}

	nd := t.Type().(*types.Named)
	var f *types.Func
	for i := 0; i < nd.NumMethods(); i++ {
		ff := nd.Method(i)
		if ff.Name() == method {
			f = ff
			break
		}
	}

	if f == nil {
		return errors.Newf("no method %s found for the type %s", method, typ)
	}

	p, err := g.m.Package("", g.path)
	if err != nil {
		return errors.Wrap(err, "set up the package renderer")
	}

	testFile := strings.TrimSuffix(g.digObjectFile(f), ".go") + "_test.go"
	r, err := p.Reuse(testFile)
	if err != nil {
		return errors.Wrap(err, "prepare test file")
	}

	if err := g.generate(p, r, f); err != nil {
		return errors.Wrap(err, "generate source code")
	}

	if err := g.m.Render(); err != nil {
		return errors.Wrap(err, "render generated source code")
	}

	return nil
}

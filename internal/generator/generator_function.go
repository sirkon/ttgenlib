package generator

import (
	"go/types"
	"strings"

	"github.com/sirkon/errors"
)

// GenerateForFunction generate table tests for a function.
func GenerateForFunction(
	pkg, fn string,
	mockLookup MockLookup,
	msgsRenderer LoggingRenderer,
	opts ...Option,
) error {
	g, err := newGenerator(pkg, mockLookup, msgsRenderer, opts...)
	if err != nil {
		return errors.Wrap(err, "init generator")
	}

	f := g.pkg.Types.Scope().Lookup(fn)
	if f == nil {
		return errors.Newf("function %s not found", fn)
	}

	s, ok := f.Type().(*types.Signature)
	if !ok {
		return errors.Newf("%s is not a function", fn)
	}

	if s.Recv() != nil {
		return errors.New("function must not be a method of any type")
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

package generator

import (
	"go/types"
	"path"

	"github.com/sirkon/errors"
	"github.com/sirkon/message"
	"golang.org/x/tools/go/packages"
)

// LocalPackage to implement PackageProvider.
func (g *Generator) LocalPackage(pkg string) (*types.Package, error) {
	pn := path.Join(g.m.Name(), pkg)
	return g.loadPackage(pn)
}

// Package to implement PackageProvider.
func (g *Generator) Package(pkg string) (*types.Package, error) {
	return g.loadPackage(pkg)
}

func (g *Generator) loadPackage(pkg string) (*types.Package, error) {
	p, ok := g.pkgs[pkg]
	if ok {
		return p.Types, nil
	}

	res, err := packages.Load(
		&packages.Config{
			Mode:    PackageLoadMode,
			Context: nil,
			Logf: func(format string, args ...interface{}) {
				message.Infof(format, args...)
			},
			Tests: false,
		},
		pkg,
	)
	if err != nil {
		return nil, errors.Wrap(err, "load package info")
	}

	for _, p := range res {
		if len(p.Errors) > 0 {
			return nil, errors.Wrap(p.Errors[0], "check returned package")
		}

		g.pkgs[p.PkgPath] = p
	}

	p, ok = g.pkgs[pkg]
	if ok {
		return p.Types, nil
	}

	return nil, ErrorPackageNotFound{pkgname: pkg}
}

var _ PackageProvider = new(Generator)

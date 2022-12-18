package generator

import (
	"go/types"
	"strings"

	"github.com/sirkon/errors"
	"github.com/sirkon/go-format"
	"github.com/sirkon/gogh"
	"github.com/sirkon/message"
)

// PackageProvider this defines a source of packages for mock lookup.
type PackageProvider interface {
	LocalPackage(path string) (*types.Package, error)
	Package(path string) (*types.Package, error)
}

// MockLookupResult a result of mock lookup.
type MockLookupResult struct {
	Name        string
	Named       *types.Named
	Type        *types.Struct
	Constructor *types.Func
}

// MockLookup is a definition of mock lookup function provided by the user.
type MockLookup func(p PackageProvider, t *types.Named) (MockLookupResult, error)

// StdMockLookup is a lookup function that should work for
// Google's [mockgen] and [pamgen] mock generators.
//  - altPaths is a list of package paths to look in if no mock was found
//    in object's own package.
//  - template is a template for mock type name based on the type name. Will
//    look be "Mock${type}" for mockgen and "${type|P}Mock" for pamgen. P is the
//    formatting option to translate original type name into the public one,
//    `pamgen` always translates mock names into public form.
//  - custom map can specify mock type names for certain types.
//
// It looks for a mock type in the given type's package first, then move to
// altPaths provided if no match was found. These criteria must be satisfied:
//   - The mock type name must be equal to template with type name applied to it.
//   - The mock type must implement the given type (it is an interface).
//   - There should be a function NewXXX(*gomock.Controller) *XXX in the package,
//     where XXX is a mock type name.
//
// [mockgen]: https://github.com/golang/mock
// [pamgen]: https://github.com/sirkon/opgen
func StdMockLookup(altPaths []string, template string, custom map[string]string) MockLookup {
	return func(p PackageProvider, t *types.Named) (res MockLookupResult, _ error) {

		var pkgs []*types.Package
		for i := 0; i < len(altPaths)+1; i++ {
			var pkgpath string
			if i == 0 {
				pkgpath = t.Obj().Pkg().Path()
			} else {
				pkgpath = altPaths[i-1]
			}

			pkg, err := p.Package(pkgpath)
			if err != nil {
				pkg, err = p.LocalPackage(pkgpath)
				if err != nil {
					message.Warning(errors.Wrapf(err, "package %s was not found", pkgpath))
					continue
				}
			}

			pkgs = append(pkgs, pkg)
		}
		if len(pkgs) == 0 {
			return res, errors.New("no package was found")
		}

		var mockName string
		if v, ok := custom[t.Obj().String()]; ok {
			mockName = v
		} else {
			mockName = format.Formatm(template, format.Values{
				"type": casesFormatter{
					value: t.Obj().Name(),
				},
			})
		}
		constructorName := "New" + mockName

		for _, pkg := range pkgs {
			res, err := mockLookup(pkg, t, mockName, constructorName)
			if err == nil {
				message.Debugf("found a mock %s for %s", res.Named, t.String())
				return res, nil
			}

			message.Warning(
				errors.Wrapf(
					err,
					"look for mock %s with constructor %s in package %s",
					mockName,
					constructorName,
					pkg.Path(),
				),
			)
		}

		return res, ErrorMockNotFound
	}
}

func mockLookup(pkg *types.Package, t *types.Named, mockName, constructor string) (res MockLookupResult, _ error) {
	// Look for mock type.
	mock := pkg.Scope().Lookup(mockName)
	if mock == nil {
		return res, ErrorMockNotFound
	}

	// Check if the pointer of the type found implements t.
	ptr := types.NewPointer(mock.Type())
	if !types.Implements(ptr, t.Underlying().(*types.Interface)) {
		return res, errors.Newf("type for mock was found but it does not implement %s", t.Obj().Name())
	}

	// Check if the mock is a structure.
	mockStruct, err := castNamedType[*types.Struct](mock.Type())
	if err != nil {
		return res, errors.Wrap(err, "get mock structure type")
	}

	// Look for the constructor function in the package.
	constr := pkg.Scope().Lookup(constructor)
	if constr == nil {
		return res, errors.Newf("type does not a have an expected constructor %a", constructor)
	}

	// It must be a function.
	s, ok := constr.Type().(*types.Signature)
	if !ok {
		return res, errors.Newf("%s is not a function", constructor)
	}

	// It must not be a method at that.
	if s.Recv() != nil {
		return res, errors.Newf("%s must not be a method", constructor)
	}

	// Must have exactly one argument.
	if s.Params().Len() != 1 {
		return res, errors.Newf("mock constructor must have exactly one argument, has %d", s.Params().Len())
	}

	// Of type *gomock.Controller
	prm, err := castNamedTypeOutOfPointer(s.Params().At(0).Type())
	if err != nil {
		return res, errors.Wrapf(
			err,
			"%s.%s type expected for the mock constructor parameter, got %s",
			gomockPath,
			gomockController,
			s.Params().At(0).Type().String(),
		)
	}
	if prm.Obj().Pkg().Path() != gomockPath || prm.Obj().Name() != gomockController {
		return res, errors.Wrapf(
			err,
			"%s.%s type expected for the mock constructor parameter, got %s",
			gomockPath,
			gomockController,
			s.Params().At(0).Type().String(),
		)
	}

	// Just one return value
	if s.Results().Len() != 1 {
		return res, errors.Newf("mock constructor must have exactly one return value, has %d", s.Results().Len())
	}

	// Being a pointer to the mock type.
	prsm, err := castNamedTypeOutOfPointer(s.Results().At(0).Type())
	if err != nil {
		return res, errors.Wrapf(
			err,
			"*%s type expected for the only result, got %s",
			mockName,
			s.Results().At(0).Type(),
		)
	}
	prmUptr := prsm.Underlying()
	if prmUptr != mockStruct {
		return res, errors.Newf(
			"*%s type expected for the only result, got %s",
			mockName,
			s.Results().At(0).Type(),
		)
	}

	res.Named = prsm
	res.Type = mockStruct
	res.Constructor = constr.(*types.Func)

	return res, nil
}

type casesFormatter struct {
	format byte
	value  string
}

// Clarify to implement format.Formatter
func (c casesFormatter) Clarify(s string) (format.Formatter, error) {
	f := strings.TrimSpace(s)
	switch f {
	case "P":
		return casesFormatter{format: f[0], value: c.value}, nil
	default:
		return nil, errors.Newf("format '%s' is not supported", f)
	}
}

// Format to implement format.Formatter
func (c casesFormatter) Format(s string) (string, error) {
	switch s {
	case "P":
		return gogh.Public(c.value), nil
	case "p":
		return gogh.Private(c.value), nil
	case "_":
		return gogh.Underscored(c.value), nil
	case "-":
		return gogh.Striked(c.value), nil
	case "R":
		return gogh.Proto(c.value), nil
	}

	return c.value, nil
}

func castNamedType[T types.Type](v types.Type) (res T, _ error) {
	vv, ok := v.(*types.Named)
	if !ok {
		return res, errors.New("must be named type")
	}

	vvv, ok := vv.Underlying().(T)
	if !ok {
		return res, errors.New("is not a structure")
	}

	return vvv, nil
}

func castNamedTypeOutOfPointer(v types.Type) (*types.Named, error) {
	vv, ok := v.(*types.Pointer)
	if !ok {
		return nil, errors.Newf("pointer to a type expected, got %T", v)
	}

	vvv, ok := vv.Elem().(*types.Named)
	if !ok {
		return nil, errors.Newf("pointer to a named type expected, got a pointer to %T", vv)
	}

	return vvv, nil
}

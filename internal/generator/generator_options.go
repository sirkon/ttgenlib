package generator

import (
	"go/types"

	"github.com/sirkon/gogh"
)

// Option generator option definition.
type Option func(g *Generator, _ optionRestriction) error

type optionRestriction struct{}

// WithNoMock adds a type which will not be mocked in tests.
func WithNoMock(pkg, name string) Option {
	return func(g *Generator, _ optionRestriction) error {
		g.nomock = append(g.nomock, doNotMock{
			path: pkg,
			name: name,
		})

		return nil
	}
}

// WithMockContext context.Context is not required to have a mock by default.
// This enables context mock requirement.
func WithMockContext(g *Generator, _ optionRestriction) error {
	var nomocks []doNotMock
	for _, nm := range g.nomock {
		if nm == contextNoMock {
			continue
		}

		nm := nm
		nomocks = append(nomocks, nm)
	}

	g.nomock = nomocks
	return nil
}

// WithMockerNames lets to set a file and type names for a mocker of a given type.
func WithMockerNames(n func(tn *types.TypeName) (fileName string, typeName string)) Option {
	return func(g *Generator, _ optionRestriction) error {
		g.mockerNames = n
		return nil
	}
}

// WithPreTest overrides default pretest renderer.
func WithPreTest(pretest func(r *gogh.GoRenderer[*gogh.Imports])) Option {
	return func(g *Generator, _ optionRestriction) error {
		g.preTest = pretest
		return nil
	}
}

// WithCtxInit overrides default context.Context init rendering.
func WithCtxInit(ctxinit func(r *gogh.GoRenderer[*gogh.Imports])) Option {
	return func(g *Generator, _ optionRestriction) error {
		g.ctxInit = ctxinit
		return nil
	}
}

// LoggingRenderer renders error messages.
// These variables:
//
//   t *testing.T
//   err error
//
// are available in the generation scope to use.
type LoggingRenderer interface {
	// ExpectedError is used to print expected error err.
	ExpectedError(r *gogh.GoRenderer[*gogh.Imports])
	// UnexpectedError used when error err was not expected.
	UnexpectedError(r *gogh.GoRenderer[*gogh.Imports])
	// ErrorWasExpected prints a message about missing error what was expected.
	ErrorWasExpected(r *gogh.GoRenderer[*gogh.Imports])
	// InvalidError prints a message about an invalid error value.
	InvalidError(r *gogh.GoRenderer[*gogh.Imports], errvar string)
}

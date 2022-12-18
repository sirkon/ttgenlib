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

// WithMockerFileName lets to set a file name for a mocker of a given type.
func WithMockerFileName(n func(*types.TypeName) string) Option {
	return func(g *Generator, _ optionRestriction) error {
		g.mockerName = n
		return nil
	}
}

// WithRenderers defines custom renderers for test processing logic.
func WithRenderers(
	preTest func(r *gogh.GoRenderer[*gogh.Imports]),
	ctxInit func(r *gogh.GoRenderer[*gogh.Imports]),
	messager MessagesRenderer,
) Option {
	return func(g *Generator, _ optionRestriction) error {
		g.preTest = preTest
		g.ctxInit = ctxInit
		g.msgr = messager

		return nil
	}
}

// MessagesRenderer this prints error processing messages.
type MessagesRenderer interface {
	// ExpectedError is used to print expected error err.
	ExpectedError(r *gogh.GoRenderer[*gogh.Imports])
	// UnexpectedError used when error err was not expected.
	UnexpectedError(r *gogh.GoRenderer[*gogh.Imports])
	// ErrorWasExpected prints a message about missing error what was expected.
	ErrorWasExpected(r *gogh.GoRenderer[*gogh.Imports])
	// InvalidError prints a message about an invalid error value.
	InvalidError(r *gogh.GoRenderer[*gogh.Imports], errvar string)
}

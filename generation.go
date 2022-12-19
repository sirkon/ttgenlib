package ttgenlib

import (
	"go/types"

	"github.com/sirkon/gogh"
	"github.com/sirkon/ttgenlib/internal/generator"
)

// GenOption code generation option type.
type GenOption = generator.Option

// GenNoMock adds an interface tye that does not need to be mocked.
// context.Context is one by default. Use GenMockContext to override
// this.
func GenNoMock(pkgname, typename string) GenOption {
	return generator.WithNoMock(pkgname, typename)
}

// GenMockContext enables mocking for context.Context parameter or field.
func GenMockContext() GenOption {
	return generator.WithMockContext
}

// GenMockerNames this option is used to override default mocker file and type name
// generation. It is <type_name>_mocker_test.go and <typeName>Mocker by default.
func GenMockerNames(
	f func(tn *types.TypeName) (filename string, typename string),
) GenOption {
	return generator.WithMockerNames(f)
}

// GenUserDefinedCodeRenderer just a shortcut for hard to read function type defintion.
type GenUserDefinedCodeRenderer = func(r *gogh.GoRenderer[*gogh.Imports])

// GenPreTest is a generator which can be used to put some additional code just after
// ctrl := ... part and right before everything else in the test logic.
func GenPreTest(pretest GenUserDefinedCodeRenderer) GenOption {
	return generator.WithPreTest(pretest)
}

// GenCtxInit overrides default context.Context initialization code rendering.
func GenCtxInit(ctxinit GenUserDefinedCodeRenderer) GenOption {
	return generator.WithCtxInit(ctxinit)
}

// GenLoggingRenderer renders error messages.
// These variables:
//
//   t *testing.T
//   err error
//
// are available in the generation scope to use.
type GenLoggingRenderer = generator.LoggingRenderer

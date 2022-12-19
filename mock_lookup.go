package ttgenlib

import "github.com/sirkon/ttgenlib/internal/generator"

// MockLookup a (functional) type to look for mocks for a given type.
type MockLookup = generator.MockLookup

// MockLookupResult a result to be returned when a mock for a given type was found.
type MockLookupResult = generator.MockLookupResult

// StandardMockLookup this is a mock lookup function that is seemingly sufficient for
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
func StandardMockLookup(altPaths []string, template string, custom map[string]string) MockLookup {
	return generator.StdMockLookup(altPaths, template, custom)
}

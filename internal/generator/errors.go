package generator

import (
	"fmt"

	"github.com/sirkon/errors"
)

// ErrorPackageNotFound package not found in the current project.
type ErrorPackageNotFound struct {
	pkgname string
}

func (e ErrorPackageNotFound) Error() string {
	return fmt.Sprintf("package '%s' not found in the current project", e.pkgname)
}

// Is to support custom handling for errors.Is.
func (e ErrorPackageNotFound) Is(err error) bool {
	_, ok := err.(ErrorPackageNotFound)
	return ok
}

// IsErrorPackageNotFound tests an error to be ErrorPackageNotFound.
func IsErrorPackageNotFound(err error) bool {
	return errors.Is(err, ErrorPackageNotFound{})
}

// ErrorMockNotFound may be returned if no mock was found.
const ErrorMockNotFound errors.Const = "mock was not found"

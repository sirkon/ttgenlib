package ttgenlib

import (
	"regexp"

	"github.com/sirkon/errors"
	"github.com/sirkon/jsonexec"
)

// goIdentifier to check correctness of text that should represent Go identifiers.
type goIdentifier string

func (id goIdentifier) String() string {
	return string(id)
}

// UnmarshalText to satisfy encoding.TestUnmarshaler.
func (id *goIdentifier) UnmarshalText(t []byte) error {
	if !goIdentifierMatcher.Match(t) {
		return errors.Newf("'%s' is invalid go identifier", string(t))
	}

	*id = goIdentifier(t)
	return nil
}

// pkgPath to check if Go package path is correct and exists.
type pkgPath string

func (p pkgPath) String() string {
	return string(p)
}

// UnmarshalText to satisfy encoding.TestUnmarshaler.
func (p *pkgPath) UnmarshalText(t []byte) error {
	var dst struct {
		ImportPath string
	}

	if err := jsonexec.Run(&dst, "go", "list", "--json", string(t)); err != nil {
		return err
	}

	*p = pkgPath(dst.ImportPath)
	return nil
}

var goIdentifierMatcher *regexp.Regexp

func init() {
	goIdentifierMatcher = regexp.MustCompile(`^[a-zA-Z_][a-zA-Z0-9_]*$`)
}

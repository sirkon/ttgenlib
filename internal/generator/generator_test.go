package generator

import (
	"testing"

	"github.com/sirkon/errors"
	"github.com/sirkon/testlog"
)

func TestNewGenerator(t *testing.T) {
	if err := GenerateForFunction(".", "newGenerator", nil); err != nil {
		testlog.Error(t, errors.Wrap(err, "create function table test template"))
	}
}

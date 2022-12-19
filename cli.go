package ttgenlib

import "github.com/willabides/kongplete"

type cliArgs struct {
	InstallCompletions kongplete.InstallCompletions `cmd:"install-completions" help:"Install completions and exit."`
	Version            versionCommand               `cmd:"" help:"Show version and exit." short:"v"`

	PkgPath  pkgPath         `help:"Package path to look in." short:"p" default:"." predict:"PKG_PATH"`
	Method   commandMethod   `cmd:"" help:"Generate test template for a method."`
	Function commandFunction `cmd:"" help:"Generate test template for a function."`
}

type runContext struct {
	args    *cliArgs
	lookup  MockLookup
	logging GenLoggingRenderer
	opts    []GenOption
}

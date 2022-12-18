package ttgenlib

type cliArgs struct {
	Version     bool `help:"Show version and exit." short:"v"`
	Completions bool `help:"Install command completions and exit." short:"c"`

	PkgPath  string `help:"Package path to look in." short:"p" default:"."`
	Type     string `arg:"" help:"Type name. Method name must be provided after it once this one is not empty."`
	Method   string `arg:"" help:"Method name. Must follow non-empty type name."`
	Function string `help:"Function name. Cannot be used together with type and method arguments." short:"f"`
}
